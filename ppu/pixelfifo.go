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
	buffer []byte
	len    int
}

func newPixelFiFo() *pixelFiFo {
	return &pixelFiFo{
		buffer: make([]byte, 16),
	}
}

func (fifo *pixelFiFo) clear() {
	fifo.len = 0
}

func (fifo *pixelFiFo) enqueue(pixData []byte) {
	for i, d := range pixData {
		fifo.buffer[i+fifo.len] = d
	}
	fifo.len += len(pixData)
}

func (fifo *pixelFiFo) setOverlay(pixData []byte, offset int) {
	for j := offset; j < len(pixData); j++ {
		p := pixData[j]
		i := j - offset

		if (fifo.buffer[i] & 0x80) != 0 {
			continue
		}
		priority := (p & 0x80) != 0
		if (priority && ((fifo.buffer[i] & 0x03) == 0)) || (!priority && (p&0x03) != 0) {
			fifo.buffer[i] = p
		}
	}
}

func (fifo *pixelFiFo) dequeue(ppu *PPU) color.Color {
	b := fifo.buffer[0]
	for i := 1; i < fifo.len; i++ {
		fifo.buffer[i-1] = fifo.buffer[i]
	}
	fifo.len--

	src := paletteSrc((b >> 4) & 0x07)
	pix := b & 0x03

	palette := ppu.bgPal
	switch src {
	case psObj0:
		palette = ppu.obj0
	case psObj1:
		palette = ppu.obj1
	}

	return palette.toColor(pix)
}
