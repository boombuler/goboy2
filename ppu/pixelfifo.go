package ppu

import (
	"image/color"
)

type paletteSrc byte

const (
	psBG paletteSrc = iota
	psObj0
	psObj1
)

type pixelFiFo struct {
}

func newPixelFiFo() *pixelFiFo {
	return new(pixelFiFo)
}

func (fifo *pixelFiFo) clear() {

}

func (fifo *pixelFiFo) len() int {
	return 0
}

func (fifo *pixelFiFo) enqueue(pixData []byte) {

}

func (fifo *pixelFiFo) dequeue(ppu *PPU) color.Color {
	var src paletteSrc
	var pix byte

	palette := ppu.bgPal
	switch src {
	case psObj0:
		palette = ppu.obj0
	case psObj1:
		palette = ppu.obj1
	}

	return palette.toColor(pix)
}
