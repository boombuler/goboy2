package ppu

import (
	"github.com/boombuler/goboy2/consts"
	"github.com/boombuler/goboy2/mmu"
)

type vramDMA struct {
	mmu    mmu.MMU
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

func newVramDMA(mmu mmu.MMU) *vramDMA {
	return &vramDMA{
		mmu: mmu,
	}
}

func (dma *vramDMA) Step(p *PPU) {
	if !dma.running {
		return
	}
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
	if dma.hdmaMode && ((!dma.hblankHandled && dma.ppuState == sHBlank) || !p.lcdEnabled()) {
		return true
	}
	return !dma.hdmaMode
}

func (dma *vramDMA) Read(addr uint16) byte {
	if dma.mmu.EmuMode() == consts.GBC {
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
	}
	return 0xFF
}

func (dma *vramDMA) Write(addr uint16, val byte) {
	if dma.mmu.EmuMode() == consts.GBC {
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
		dma.hblankHandled = false
		dma.ppuState = sOAMRead // Any other then HBlank...
	}
}
