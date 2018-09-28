package ppu

import (
	"image/color"
	"testing"
)

func TestGBCPalette(t *testing.T) {

	testCases := []struct {
		Lo    byte
		Hi    byte
		Color color.RGBA
	}{
		{0xE0, 0x03, color.RGBA{0, 0x1F, 0, 0}},
		{0x1F, 0x7C, color.RGBA{0x1F, 0, 0x1F, 0}},
		{0xFF, 0x03, color.RGBA{0x1F, 0x1F, 0, 0}},
		{0x97, 0xDA, color.RGBA{0x17, 0x14, 0x16, 0}},
		{0xD6, 0xB0, color.RGBA{0x16, 0x06, 0x0C, 0}},
		{0xEB, 0x9B, color.RGBA{0x0B, 0x1F, 0x06, 0}},
	}

	p := newGBCPalette(0)
	p.Write(0, 0x80)

	for i, tc := range testCases {
		p.Write(1, tc.Lo)
		p.Write(1, tc.Hi)

		col := p.toColor(i/4, byte(i%4)).(color.RGBA)
		col.R /= 8
		col.G /= 8
		col.B /= 8
		if col.R != tc.Color.R || col.G != tc.Color.G || col.B != tc.Color.B {
			t.Errorf("Failed: %02X%02X Expected %02X%02X%02X but got %02X%02X%02X", tc.Lo, tc.Hi, tc.Color.R, tc.Color.G, tc.Color.B, col.R, col.G, col.B)
		}
	}
}
