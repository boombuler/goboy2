package ppu

import (
	"image/color"
)

// Fifo Format: BPPP _FCC (B = BG-Palette; P = Palette; F = Priority-Flag; C = Color)

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
	return pixel&0x04 != 0
}
func useBGPal(pixel byte) bool {
	return pixel&0x80 != 0
}
func palIdx(pixel byte) int {
	return int(pixel>>4) & 0x07
}

func (fifo *pixelFiFo) setOverlay(pixData []byte, offset int) {
	for j := offset; j < len(pixData); j++ {
		p := pixData[j]
		i := j - offset
		bi := fifo.idx(i)
		if !useBGPal(fifo.buffer[bi]) || colIdx(p) == 0 {
			continue
		}
		//bgPrio :=  prio(fifo.buffer[bi])

		priority := prio(p) //|| bgPrio

		if (priority && (colIdx(fifo.buffer[bi]) == 0)) || (!priority && (colIdx(p) != 0)) {
			fifo.buffer[bi] = p
		}
	}
}

func (fifo *pixelFiFo) dequeue(ppu *PPU) color.Color {
	b := fifo.buffer[fifo.idx(0)]
	fifo.startIdx = (fifo.startIdx + 1) % len(fifo.buffer)
	fifo.len--

	pix := colIdx(b)

	var palette palette
	if ppu.mmu.GBC() {
		palette = ppu.obcPal
		if useBGPal(b) {
			palette = ppu.bgcPal
		}
	} else {
		palette = ppu.objPal
		if useBGPal(b) {
			palette = ppu.bgPal
		}
	}
	return palette.toColor(palIdx(b), pix)
}
