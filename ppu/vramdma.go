package ppu

import "goboy2/consts"

type vramDMA struct {
	hdma12 uint16
	hdma34 uint16

	src      uint16
	dest     uint16
	length   byte
	running  bool
	hdmaMode bool

	ppuState      ppuState
	hblankHandled bool
	timer         int
}

func (dma *vramDMA) Step(p *PPU) {
	if ppus := p.state(); dma.ppuState != p.state() {
		if ppus == sHBlank {
			dma.hblankHandled = false
		}
		dma.ppuState = ppus
	}

	if !dma.active(p) {
		return
	}

	if dma.timer++; dma.timer < 0x20 {
		return
	}
	dma.timer = 0
	for i := uint16(0); i < 0x10; i++ {
		p.mmu.Write(dma.dest, p.mmu.Read(dma.src))
		dma.src++
		dma.dest++
	}
	dma.length--
	dma.hblankHandled = true

	if dma.length == 0 {
		dma.running = false
		dma.length = 0x7F
	}
}

func (dma *vramDMA) active(p *PPU) bool {
	if !dma.running {
		return false
	} else if dma.hdmaMode && ((!dma.hblankHandled && dma.ppuState == sHBlank) || !p.lcdEnabled()) {
		return true
	}
	return !dma.hdmaMode
}

func (dma *vramDMA) Read(addr uint16) byte {
	switch addr {
	case consts.AddrHDMA1:
		return byte(dma.hdma12 >> 8)
	case consts.AddrHDMA2:
		return byte(dma.hdma12)
	case consts.AddrHDMA3:
		return byte(dma.hdma34 >> 8)
	case consts.AddrHDMA4:
		return byte(dma.hdma34)
	case consts.AddrHDMA5:
		r := byte(0x00)
		if dma.running {
			r = 1 << 7
		}
		return r | dma.length
	}
	return 0xFF
}

func (dma *vramDMA) Write(addr uint16, val byte) {
	switch addr {
	case consts.AddrHDMA1:
		dma.hdma12 = dma.hdma12&0x00FF | (uint16(val) << 8)
	case consts.AddrHDMA2:
		dma.hdma12 = dma.hdma12&0xFF00 | uint16(val)
	case consts.AddrHDMA3:
		dma.hdma34 = dma.hdma34&0x00FF | (uint16(val) << 8)
	case consts.AddrHDMA4:
		dma.hdma34 = dma.hdma34&0xFF00 | uint16(val)
	case consts.AddrHDMA5:
		dma.startTransfer(val)
	}
}

func (dma *vramDMA) startTransfer(ctrl byte) {
	dma.length = ctrl & 0x7F

	ctrlBit := ctrl&0x80 != 0
	if !ctrlBit && dma.hdmaMode {
		dma.running = false
	} else {
		dma.running = true
		dma.hdmaMode = ctrlBit
		dma.src = dma.hdma12 & 0xFFF0
		dma.dest = (dma.hdma34 & 0x1FF0) | 0x8000
		dma.timer = 0
	}
}
