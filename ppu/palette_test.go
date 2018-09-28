package ppu

import (
	"testing"
)

func TestGBCPalette(t *testing.T) {
	testCases := []struct {
		Hi byte
		Lo byte

		Color RGB
	}{
		{0x03, 0xE0, RGB{0, 0x1F, 0}},
		{0x7C, 0x1F, RGB{0x1F, 0, 0x1F}},
		{0x03, 0xFF, RGB{0x1F, 0x1F, 0}},
		{0xDA, 0x97, RGB{0x17, 0x14, 0x16}},
		{0xB0, 0xD6, RGB{0x16, 0x06, 0x0C}},
		{0x9B, 0xEB, RGB{0x0B, 0x1F, 0x06}},
	}

	p := newGBCPalette(0)
	p.Write(0, 0x80)
	failed := false
	for i, tc := range testCases {
		p.Write(1, tc.Lo)
		p.Write(1, tc.Hi)

		col := p.toColor(i/4, byte(i%4))
		col.R = col.R >> colorShift
		col.G = col.G >> colorShift
		col.B = col.B >> colorShift
		if col.R != tc.Color.R || col.G != tc.Color.G || col.B != tc.Color.B {
			t.Errorf("Failed: %02X%02X Expected %02X%02X%02X but got %02X%02X%02X", tc.Hi, tc.Lo, tc.Color.R, tc.Color.G, tc.Color.B, col.R, col.G, col.B)
			failed = true
		}
	}

	if failed {
		p.PrintRam()
	}
}
