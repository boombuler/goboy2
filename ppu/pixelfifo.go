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
	buffer   []byte
	len      int
	startIdx int
}

func newPixelFiFo() *pixelFiFo {
	return &pixelFiFo{
		buffer: make([]byte, 16),
	}
}

func (fifo *pixelFiFo) clear() {
	fifo.len = 0
}

func (fifo *pixelFiFo) idx(i int) int {
	return (i + fifo.startIdx) % len(fifo.buffer)
}

func (fifo *pixelFiFo) enqueue(pixData []byte) {
	for i, d := range pixData {
		fifo.buffer[fifo.idx(i+fifo.len)] = d
	}
	fifo.len += len(pixData)
}

func colIdx(pixel byte) byte {
	return pixel & 0x03
}
func prio(pixel byte) bool {
	return pixel&0x80 != 0
}

func (fifo *pixelFiFo) setOverlay(pixData []byte, offset int) {
	for j := offset; j < len(pixData); j++ {
		p := pixData[j]
		i := j - offset
		bi := fifo.idx(i)
		if ps := paletteSrc((fifo.buffer[bi] >> 4) & 0x07); ps != psBG {
			continue
		}
		priority := prio(p)
		if (priority && (colIdx(fifo.buffer[bi]) == 0)) || (!priority && (colIdx(p) != 0)) {
			fifo.buffer[bi] = p
		}
	}
}

func (fifo *pixelFiFo) dequeue(ppu *PPU) color.Color {
	b := fifo.buffer[fifo.idx(0)]
	fifo.startIdx = (fifo.startIdx + 1) % len(fifo.buffer)
	fifo.len--

	src := paletteSrc((b >> 4) & 0x07)
	pix := colIdx(b)

	palette := ppu.bgPal
	switch src {
	case psObj0:
		palette = ppu.obj0
	case psObj1:
		palette = ppu.obj1
	}

	return palette.toColor(pix)
}
