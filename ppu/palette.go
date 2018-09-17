package ppu

import (
	"image"
	"image/color"
)

type palette byte

var gbColors = []color.Color{
	color.Gray{0xEB},
	color.Gray{0xC4},
	color.Gray{0x60},
	color.Gray{0x00},
}

const DISPLAY_WIDTH int = 160
const DISPLAY_HEIGHT int = 144

var emptyScreen = newScreen()

func newScreen() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, DISPLAY_WIDTH, DISPLAY_HEIGHT))
}

func (p palette) toColor(val byte) color.Color {
	return gbColors[0x03&(p>>((val&0x03)<<1))]
}
