package ppu

import (
	"goboy2/mmu"
)

type hblank struct {
	ticks uint16
}

func (hb *hblank) state() ppuState {
	return sHBlank
}
func (hb *hblank) step(ppu *PPU) bool {
	hb.ticks++
	return hb.ticks >= 456
}
func (hb *hblank) start(ppu *PPU) bool {
	if int(ppu.ly) >= DisplayHeight {
		return false
	}
	hb.ticks = ppu.ticksInLine
	ppu.requstLcdcInterrupt(liHBlank)
	return true
}

type vblank struct {
	ticks uint16
}

func (vb *vblank) state() ppuState {
	return sVBlank
}
func (vb *vblank) step(ppu *PPU) bool {
	vb.ticks++
	return vb.ticks >= 456
}
func (vb *vblank) start(ppu *PPU) bool {
	if int(ppu.ly) < DisplayHeight {
		return false
	}
	vb.ticks = 0
	if int(ppu.ly) == DisplayHeight {
		ppu.screenOut <- ppu.curScreen
		ppu.curScreen = newScreen()
		ppu.requstLcdcInterrupt(liVBlank)
		ppu.mmu.RequestInterrupt(mmu.IRQVBlank)
	}
	return true
}
