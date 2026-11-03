package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
)

// AgentState represents the operational and hardware status of the biometric agent.
type AgentState int

const (
	StateReady AgentState = iota
	StateCapturing
	StateDisconnected
	StateError
)

type iconTheme struct {
	RidgeColor      color.RGBA
	CoreGlowColor   color.RGBA
	ScanBeamColor   color.RGBA
	ShowScanBeam    bool
	BadgeColor      color.RGBA
	ShowBadge       bool
}

var stateThemes = map[AgentState]iconTheme{
	StateReady: {
		// Vibrant Emerald / Cyber Cyan - indicates live ready biometric sensor
		RidgeColor:    color.RGBA{R: 0x00, G: 0xE5, B: 0xB0, A: 0xFF}, // #00E5B0
		CoreGlowColor: color.RGBA{R: 0x00, G: 0xB4, B: 0xD8, A: 0xFF}, // #00B4D8
		ShowScanBeam:  false,
		ShowBadge:     false,
	},
	StateCapturing: {
		// Radiant Electric Amber / Gold - indicates finger down / minutiae reading in progress
		RidgeColor:    color.RGBA{R: 0xFF, G: 0xB7, B: 0x03, A: 0xFF}, // #FFB703
		CoreGlowColor: color.RGBA{R: 0xFB, G: 0x85, B: 0x00, A: 0xFF}, // #FB8500
		ScanBeamColor: color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xF0}, // White-hot beam
		ShowScanBeam:  true,
		ShowBadge:     false,
	},
	StateDisconnected: {
		// Elegant Muted Slate Gray - device unplugged or waiting
		RidgeColor:    color.RGBA{R: 0x8A, G: 0x9B, B: 0xA8, A: 0xCC}, // Slate
		CoreGlowColor: color.RGBA{R: 0x5C, G: 0x70, B: 0x80, A: 0x99},
		ShowScanBeam:  false,
		BadgeColor:    color.RGBA{R: 0xFA, G: 0xBF, B: 0x00, A: 0xFF}, // Amber unplugged indicator
		ShowBadge:     true,
	},
	StateError: {
		// Vivid Crimson Coral - driver missing, USB error, or service stopped
		RidgeColor:    color.RGBA{R: 0xFF, G: 0x47, B: 0x57, A: 0xFF}, // #FF4757
		CoreGlowColor: color.RGBA{R: 0xED, G: 0x29, B: 0x39, A: 0xFF},
		ShowScanBeam:  false,
		BadgeColor:    color.RGBA{R: 0xFF, G: 0x00, B: 0x33, A: 0xFF},
		ShowBadge:     true,
	},
}

// GenerateBiometricIcon generates an elegant, high-DPI biometric fingerprint icon
// dynamically styled for the specified AgentState, returned as valid Windows ICO bytes.
func GenerateBiometricIcon(state AgentState, size int) []byte {
	if size <= 0 {
		size = 64
	}
	theme, ok := stateThemes[state]
	if !ok {
		theme = stateThemes[StateReady]
	}

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// Clear to transparent
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.Transparent}, image.Point{}, draw.Src)

	cx := float64(size) / 2.0
	cy := float64(size) * 0.52

	// Render concentric stylized fingerprint ridges
	// An authentic biometric loop consists of several concentric rounded archs/loops
	// with varying radii and subtle gaps mimicking human friction ridges.
	ridges := []struct {
		rx, ry float64 // horizontal & vertical radii
		startA float64 // start angle in radians
		endA   float64 // end angle in radians
		weight float64 // thickness in pixels
	}{
		// Innermost core whorl
		{rx: float64(size) * 0.10, ry: float64(size) * 0.12, startA: -math.Pi * 0.85, endA: math.Pi * 0.85, weight: float64(size) * 0.055},
		// Second loop
		{rx: float64(size) * 0.18, ry: float64(size) * 0.22, startA: -math.Pi * 0.90, endA: math.Pi * 0.90, weight: float64(size) * 0.055},
		// Third loop
		{rx: float64(size) * 0.26, ry: float64(size) * 0.31, startA: -math.Pi * 0.92, endA: math.Pi * 0.92, weight: float64(size) * 0.055},
		// Fourth outer arch
		{rx: float64(size) * 0.34, ry: float64(size) * 0.40, startA: -math.Pi * 0.88, endA: math.Pi * 0.88, weight: float64(size) * 0.055},
		// Outermost boundary contour
		{rx: float64(size) * 0.42, ry: float64(size) * 0.48, startA: -math.Pi * 0.75, endA: math.Pi * 0.75, weight: float64(size) * 0.050},
	}

	for _, ridge := range ridges {
		drawEllipticalArc(img, cx, cy, ridge.rx, ridge.ry, ridge.startA, ridge.endA, ridge.weight, theme.RidgeColor)
	}

	// Draw central biometric core dot
	coreR := float64(size) * 0.04
	drawCircleFilled(img, cx, cy+float64(size)*0.02, coreR, theme.CoreGlowColor)

	// If capturing, overlay dynamic laser/scanning light beam across the center
	if theme.ShowScanBeam {
		beamY := cy + math.Sin(float64(size))*float64(size)*0.1
		beamH := float64(size) * 0.08
		for y := int(beamY - beamH/2); y <= int(beamY+beamH/2); y++ {
			if y < 0 || y >= size {
				continue
			}
			dist := math.Abs(float64(y) - beamY)
			alpha := uint8(float64(theme.ScanBeamColor.A) * (1.0 - dist/(beamH/2)))
			for x := int(cx - float64(size)*0.45); x <= int(cx+float64(size)*0.45); x++ {
				if x < 0 || x >= size {
					continue
				}
				orig := img.RGBAAt(x, y)
				blended := blendOver(orig, color.RGBA{R: theme.ScanBeamColor.R, G: theme.ScanBeamColor.G, B: theme.ScanBeamColor.B, A: alpha})
				img.SetRGBA(x, y, blended)
			}
		}
	}

	// If disconnected or error, draw an alert badge in the bottom-right corner
	if theme.ShowBadge {
		bx := float64(size) * 0.80
		by := float64(size) * 0.80
		br := float64(size) * 0.16
		// Black ring border around badge for crisp separation
		drawCircleFilled(img, bx, by, br+2, color.RGBA{R: 0x10, G: 0x14, B: 0x1A, A: 0xFF})
		drawCircleFilled(img, bx, by, br, theme.BadgeColor)
		// Small white icon in badge center
		drawCircleFilled(img, bx, by, br*0.4, color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})
	}

	return encodeICOFromImage(img)
}

// GenerateBiometricPNG generates the raw PNG bytes for the state.
func GenerateBiometricPNG(state AgentState, size int) []byte {
	ico := GenerateBiometricIcon(state, size)
	if len(ico) > 22 {
		return ico[22:] // PNG payload begins at offset 22
	}
	return nil
}

func drawEllipticalArc(img *image.RGBA, cx, cy, rx, ry, startA, endA, weight float64, c color.RGBA) {
	steps := int((rx + ry) * 3)
	if steps < 60 {
		steps = 60
	}
	stepA := (endA - startA) / float64(steps)

	halfW := weight / 2.0
	for i := 0; i <= steps; i++ {
		a := startA + float64(i)*stepA
		px := cx + rx*math.Sin(a)
		py := cy - ry*math.Cos(a)

		// Stamp soft circle brush
		minX := int(math.Floor(px - halfW))
		maxX := int(math.Ceil(px + halfW))
		minY := int(math.Floor(py - halfW))
		maxY := int(math.Ceil(py + halfW))

		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
					continue
				}
				d := math.Hypot(float64(x)-px, float64(y)-py)
				if d <= halfW {
					alpha := float64(c.A)
					if d > halfW-1.0 {
						alpha *= (halfW - d) // anti-aliasing edge
					}
					stamp := color.RGBA{R: c.R, G: c.G, B: c.B, A: uint8(alpha)}
					orig := img.RGBAAt(x, y)
					img.SetRGBA(x, y, blendOver(orig, stamp))
				}
			}
		}
	}
}

func drawCircleFilled(img *image.RGBA, cx, cy, r float64, c color.RGBA) {
	minX := int(math.Floor(cx - r - 1))
	maxX := int(math.Ceil(cx + r + 1))
	minY := int(math.Floor(cy - r - 1))
	maxY := int(math.Ceil(cy + r + 1))

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
				continue
			}
			d := math.Hypot(float64(x)-cx, float64(y)-cy)
			if d <= r {
				alpha := float64(c.A)
				if d > r-1.0 {
					alpha *= (r - d)
				}
				stamp := color.RGBA{R: c.R, G: c.G, B: c.B, A: uint8(alpha)}
				orig := img.RGBAAt(x, y)
				img.SetRGBA(x, y, blendOver(orig, stamp))
			}
		}
	}
}

func blendOver(dst, src color.RGBA) color.RGBA {
	if src.A == 0 {
		return dst
	}
	if src.A == 255 || dst.A == 0 {
		return src
	}
	sa := float64(src.A) / 255.0
	da := float64(dst.A) / 255.0
	outA := sa + da*(1.0-sa)
	if outA <= 0 {
		return color.RGBA{}
	}
	r := (float64(src.R)*sa + float64(dst.R)*da*(1.0-sa)) / outA
	g := (float64(src.G)*sa + float64(dst.G)*da*(1.0-sa)) / outA
	b := (float64(src.B)*sa + float64(dst.B)*da*(1.0-sa)) / outA
	return color.RGBA{
		R: uint8(math.Round(r)),
		G: uint8(math.Round(g)),
		B: uint8(math.Round(b)),
		A: uint8(math.Round(outA * 255.0)),
	}
}

// encodeICOFromImage wraps PNG-encoded image bytes in a Windows ICO structure.
func encodeICOFromImage(img image.Image) []byte {
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return nil
	}
	pngData := pngBuf.Bytes()

	var ico bytes.Buffer
	// ICONDIR Header (6 bytes)
	_ = binary.Write(&ico, binary.LittleEndian, uint16(0)) // Reserved = 0
	_ = binary.Write(&ico, binary.LittleEndian, uint16(1)) // Type = 1 (Icon)
	_ = binary.Write(&ico, binary.LittleEndian, uint16(1)) // Image Count = 1

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	if w >= 256 {
		w = 0
	}
	if h >= 256 {
		h = 0
	}

	// ICONDIRENTRY (16 bytes)
	ico.WriteByte(byte(w))
	ico.WriteByte(byte(h))
	ico.WriteByte(0) // Color count (0 for 32-bit RGBA)
	ico.WriteByte(0) // Reserved
	_ = binary.Write(&ico, binary.LittleEndian, uint16(1))  // Planes = 1
	_ = binary.Write(&ico, binary.LittleEndian, uint16(32)) // Bit count = 32
	_ = binary.Write(&ico, binary.LittleEndian, uint32(len(pngData))) // Size
	_ = binary.Write(&ico, binary.LittleEndian, uint32(22))           // Offset

	// Write PNG bytes
	ico.Write(pngData)
	return ico.Bytes()
}

