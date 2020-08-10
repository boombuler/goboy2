package ppu

import (
	"github.com/boombuler/goboy2/consts"
	"image"
	"image/color"
)

type ScreenImage [consts.DisplayHeight * consts.DisplayWidth]RGB

type RGB struct{ R, G, B byte }

func (c RGB) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = 0xFFFF
	return
}

var _ image.Image = &ScreenImage{}

var rgbColorModel = color.ModelFunc(rgbModel)

func rgbModel(c color.Color) color.Color {
	if _, ok := c.(RGB); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	return RGB{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
}

func (s *ScreenImage) ColorModel() color.Model {
	return rgbColorModel
}

func (s *ScreenImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, consts.DisplayWidth, consts.DisplayHeight)
}

func (s *ScreenImage) At(x, y int) color.Color {
	return s[y*consts.DisplayWidth+x]
}
