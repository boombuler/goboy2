package ppu

import (
	"image"
	"image/color"
)

type palette byte

var gbColors = []color.Color{
	color.RGBAModel.Convert(color.Gray{0xEB}),
	color.RGBAModel.Convert(color.Gray{0xC4}),
	color.RGBAModel.Convert(color.Gray{0x60}),
	color.RGBAModel.Convert(color.Gray{0x00}),
}

var emptyScreen = newScreen()

func newScreen() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, DisplayWidth, DisplayHeight))
}

func (p palette) toColor(val byte) color.Color {
	shift := (val & 0x03) * 2
	color := (p >> shift)
	return gbColors[0x03&color]
}
