package ppu

// Fifo Format: BPPP _FCC (B = BG-Palette; P = Palette; F = Priority-Flag; C = Color)

type pixelFiFo struct {
	buffer   []byte
	oamIdx   []int
	len      int
	startIdx int
}

const fifoBufferLen = 16

func newPixelFiFo() *pixelFiFo {
	return &pixelFiFo{
		buffer: make([]byte, fifoBufferLen),
		oamIdx: make([]int, fifoBufferLen),
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
		idx := fifo.idx(i + fifo.len)
		fifo.buffer[idx] = d
		fifo.oamIdx[idx] = -1
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

func (fifo *pixelFiFo) setOverlay(ppu *PPU, pixData []byte, offset int, oamIndex int) {
	for j := offset; j < len(pixData); j++ {
		p := pixData[j]
		i := j - offset
		bi := fifo.idx(i)
		if ppu.usePixel(fifo.buffer[bi], p, fifo.oamIdx[bi], oamIndex) {
			fifo.buffer[bi] = p
			fifo.oamIdx[bi] = oamIndex
		}
	}
}

func (ppu *PPU) usePixel(curPix, newPix byte, curOAMIdx, newOAMIdx int) bool {
	if colIdx(newPix) == 0 {
		return false
	}

	if ppu.dmgMode() {
		if !useBGPal(curPix) {
			return false
		}
		return !prio(newPix) || colIdx(curPix) == 0
	}
	// GBC Mode
	if useBGPal(curPix) {
		if !ppu.masterPriority() {
			return true
		} else if prio(curPix) {
			return colIdx(curPix) == 0
		}
		return !prio(newPix) || colIdx(curPix) == 0
	}
	return curOAMIdx > newOAMIdx
}

func (fifo *pixelFiFo) dequeue(ppu *PPU) RGB {
	b := fifo.buffer[fifo.idx(0)]
	fifo.startIdx = (fifo.startIdx + 1) % fifoBufferLen
	fifo.len--

	pix := colIdx(b)

	var palette palette
	if ppu.gbc {
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
