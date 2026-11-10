// Package printing provides ESC/POS command helpers and printer drivers
// for thermal receipt printers (e.g. P-25, Epson TM series).
//
// Build constraint: printer drivers (USB, Serial) are Windows-only.
// Network printing works on all platforms.
package printing

import "strings"

// ─── ESC/POS byte constants ───────────────────────────────────────────────────

var (
	ESC = byte(0x1B)
	GS  = byte(0x1D)
	LF  = byte(0x0A)
	CR  = byte(0x0D)
	HT  = byte(0x09)
)

// COLS is the number of printable columns at standard font for 80 mm paper.
const COLS = 48

// ─── Command builders ─────────────────────────────────────────────────────────

// CmdInit resets the printer to default settings.
func CmdInit() []byte { return []byte{ESC, '@'} }

// CmdCut sends a partial paper cut command.
func CmdCut() []byte { return []byte{GS, 'V', 66, 0} }

// CmdFeedLines advances paper by n lines.
func CmdFeedLines(n int) []byte { return []byte{ESC, 'd', byte(n)} }

// CmdAlign sets text alignment: "left", "center", or "right".
func CmdAlign(a string) []byte {
	switch a {
	case "center":
		return []byte{ESC, 'a', 1}
	case "right":
		return []byte{ESC, 'a', 2}
	default:
		return []byte{ESC, 'a', 0}
	}
}

// CmdBold enables or disables bold text.
func CmdBold(on bool) []byte {
	if on {
		return []byte{ESC, 'E', 1}
	}
	return []byte{ESC, 'E', 0}
}

// CmdDoubleHeight enables or disables double-height text.
func CmdDoubleHeight(on bool) []byte {
	if on {
		return []byte{GS, '!', 0x01}
	}
	return []byte{GS, '!', 0x00}
}

// CmdDoubleSize enables or disables double-width + double-height text.
func CmdDoubleSize(on bool) []byte {
	if on {
		return []byte{GS, '!', 0x11}
	}
	return []byte{GS, '!', 0x00}
}

// CmdUnderline enables or disables underline.
func CmdUnderline(on bool) []byte {
	if on {
		return []byte{ESC, '-', 1}
	}
	return []byte{ESC, '-', 0}
}

// CmdBarcodeHeight sets the barcode height in dots.
func CmdBarcodeHeight(h byte) []byte { return []byte{GS, 'h', h} }

// CmdBarcodeWidth sets the barcode module width.
func CmdBarcodeWidth(w byte) []byte { return []byte{GS, 'w', w} }

// CmdBarcodeHRI sets the HRI (human-readable interpretation) position.
// pos 0=none, 1=above, 2=below, 3=both
func CmdBarcodeHRI(pos byte) []byte { return []byte{GS, 'H', pos} }

// CmdBarcode sends a CODE128 barcode (System B, m=73).
func CmdBarcode(data string) []byte {
	b := []byte{GS, 'k', 73, byte(len(data))}
	b = append(b, []byte(data)...)
	return b
}

// CmdBarcode39 sends a CODE39 barcode (System B, m=69).
func CmdBarcode39(data string) []byte {
	b := []byte{GS, 'k', 69, byte(len(data))}
	b = append(b, []byte(data)...)
	return b
}

// ─── Layout helpers ───────────────────────────────────────────────────────────

// Divider returns a full-width dashed separator line.
func Divider() []byte {
	return []byte(strings.Repeat("-", COLS) + "\n")
}

// PadRight pads s with spaces on the right to width w.
func PadRight(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

// PadLeft pads s with spaces on the left to width w.
func PadLeft(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return strings.Repeat(" ", w-len(s)) + s
}

// Col2 returns a two-column line: left-aligned + right-aligned within width.
func Col2(left, right string, width int) []byte {
	space := width - len(left) - len(right)
	if space < 1 {
		space = 1
	}
	line := left + strings.Repeat(" ", space) + right + "\n"
	return []byte(line)
}
