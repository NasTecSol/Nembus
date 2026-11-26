package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
)

type StatusKind int

const (
	StatusStandby StatusKind = iota
	StatusActive
	StatusWarning
	StatusUpdate
)

// GenerateStatusIcon produces a valid Windows .ICO (PNG-encoded 32x32) with a
// crisp circular Nembus SAP Agent badge colored by the current health state.
func GenerateStatusIcon(kind StatusKind) []byte {
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	var fill color.RGBA
	switch kind {
	case StatusActive:
		fill = color.RGBA{R: 16, G: 185, B: 129, A: 255} // Emerald green
	case StatusWarning:
		fill = color.RGBA{R: 245, G: 158, B: 11, A: 255} // Amber
	case StatusUpdate:
		fill = color.RGBA{R: 59, G: 130, B: 246, A: 255} // Blue
	default:
		fill = color.RGBA{R: 100, G: 116, B: 139, A: 255} // Slate standby
	}
	ring := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	cx, cy := float64(size)/2.0, float64(size)/2.0
	outerR := 14.5
	innerR := 11.5

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			d := math.Hypot(dx, dy)
			if d <= innerR {
				img.SetRGBA(x, y, fill)
			} else if d <= outerR {
				img.SetRGBA(x, y, ring)
			}
		}
	}

	var pngBuf bytes.Buffer
	_ = png.Encode(&pngBuf, img)
	pngBytes := pngBuf.Bytes()

	// Wrap PNG bytes in a standard 1-image ICO container (Windows Vista+ native format).
	var ico bytes.Buffer
	_ = binary.Write(&ico, binary.LittleEndian, uint16(0)) // reserved
	_ = binary.Write(&ico, binary.LittleEndian, uint16(1)) // type: 1 = ICO
	_ = binary.Write(&ico, binary.LittleEndian, uint16(1)) // count: 1 image

	// ICONDIRENTRY (16 bytes)
	ico.WriteByte(byte(size))                              // width
	ico.WriteByte(byte(size))                              // height
	ico.WriteByte(0)                                       // color count
	ico.WriteByte(0)                                       // reserved
	_ = binary.Write(&ico, binary.LittleEndian, uint16(1)) // color planes
	_ = binary.Write(&ico, binary.LittleEndian, uint16(32))
	_ = binary.Write(&ico, binary.LittleEndian, uint32(len(pngBytes)))
	_ = binary.Write(&ico, binary.LittleEndian, uint32(6+16)) // offset
	ico.Write(pngBytes)

	return ico.Bytes()
}
