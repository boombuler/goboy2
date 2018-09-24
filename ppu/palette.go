package ppu

import (
	"image"
	"image/color"
	"sync"
)

type palette byte

var gbColors = []color.Color{
	color.RGBAModel.Convert(color.Gray{0xEB}),
	color.RGBAModel.Convert(color.Gray{0xC4}),
	color.RGBAModel.Convert(color.Gray{0x60}),
	color.RGBAModel.Convert(color.Gray{0x00}),
}

var emptyScreen = newScreen()

var screenPool = &sync.Pool{
	New: func() interface{} {
		return image.NewRGBA(image.Rect(0, 0, DisplayWidth, DisplayHeight))
	},
}

func FreeScreen(img *image.RGBA) {
	screenPool.Put(img)
}

func dropFrames(output chan<- *image.RGBA) chan<- *image.RGBA {
	input := make(chan *image.RGBA)

	go func() {
		lastImg := <-input
		for {
			select {
			case img := <-input:
				FreeScreen(lastImg)
				lastImg = img
			case output <- lastImg:
			}
		}
	}()

	return input
}

func newScreen() *image.RGBA {
	return screenPool.Get().(*image.RGBA)
}

func (p palette) toColor(val byte) color.Color {
	shift := (val & 0x03) * 2
	color := (p >> shift)
	return gbColors[0x03&color]
}
