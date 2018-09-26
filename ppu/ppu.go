package ppu

import (
	"goboy2/consts"
	"goboy2/mmu"
	"image"
)

type PPU struct {
	mmu            mmu.MMU
	phaseIdx       int
	phases         []ppuPhase
	visibleSprites []*spriteData
	ticksInLine    uint16
	dma            *vramDMA

	lcdc    byte
	ie      lcdInterrupts
	ly      byte
	lyc     byte
	scrollY byte
	scrollX byte
	winX    byte
	winY    byte
	bgPal   gbPalette
	objPal  gbPalette
	bgcPal  *gbcPalette
	obcPal  *gbcPalette

	screenOut chan<- *image.RGBA
	curScreen *image.RGBA

	vram0  vRAM
	vram1  vRAM
	vramHi bool
	oam    *oam
}

// New creates a new ppu and connects it to the given mmu
func New(mmu mmu.MMU, screen chan<- *image.RGBA) *PPU {
	gbc := mmu.GBC()
	ppu := &PPU{
		mmu:       mmu,
		vram0:     newVRAM(),
		oam:       newOAM(gbc),
		screenOut: screen,
		phaseIdx:  0,
		phases: []ppuPhase{
			new(oamSearch),
			newPixelTransfer(),
			new(hblank),
			new(vblank),
		},
	}
	mmu.ConnectPPU(ppu)
	mmu.AddIODevice(ppu, consts.AddrLCDC, consts.AddrSTAT, consts.AddrSCROLLY, consts.AddrSCROLLX,
		consts.AddrLY, consts.AddrLYC, consts.AddrBGP, consts.AddrOBJECTPALETTE0, consts.AddrOBJECTPALETTE1,
		consts.AddrWY, consts.AddrWX)
	if gbc {
		ppu.vram1 = newVRAM()
		ppu.dma = new(vramDMA)
		ppu.bgcPal = &gbcPalette{IndexAdr: consts.AddrBGPI}
		ppu.obcPal = &gbcPalette{IndexAdr: consts.AddrOBPI}
		mmu.AddIODevice(ppu.dma, consts.AddrHDMA1, consts.AddrHDMA2, consts.AddrHDMA3, consts.AddrHDMA4, consts.AddrHDMA5)
		mmu.AddIODevice(ppu.bgcPal, consts.AddrBGPI, consts.AddrBGPD)
		mmu.AddIODevice(ppu.obcPal, consts.AddrOBPI, consts.AddrOBPD)
		mmu.AddIODevice(ppu, consts.AddrVBK)
	}
	return ppu
}

func (p *PPU) state() ppuState {
	return p.phases[p.phaseIdx].state()
}

func (p *PPU) Read(addr uint16) byte {
	switch addr {
	case consts.AddrVBK:
		if p.vramHi {
			return 0xFF
		}
		return 0xFE
	case consts.AddrLCDC:
		return p.lcdc
	case consts.AddrSTAT:
		var coincidence byte
		if p.ly == p.lyc {
			coincidence = 1
		}
		return byte(p.ie) | (coincidence << 2) | (byte(p.state()) & 0x03) | 0x80
	case consts.AddrSCROLLY:
		return p.scrollY
	case consts.AddrSCROLLX:
		return p.scrollX
	case consts.AddrLY:
		return p.ly
	case consts.AddrLYC:
		return p.lyc
	case consts.AddrBGP:
		return byte(p.bgPal)
	case consts.AddrOBJECTPALETTE0:
		return byte(p.objPal)
	case consts.AddrOBJECTPALETTE1:
		return byte(p.objPal >> 8)
	case consts.AddrWY:
		return p.winY
	case consts.AddrWX:
		return p.winX
	default:
		if addr >= 0x8000 && addr <= 0x9FFF {
			if p.state().canAccessVRAM() {
				if p.mmu.GBC() && p.vramHi {
					return p.vram1.Read(addr)
				}
				return p.vram0.Read(addr)
			}
			return 0xFF
		} else if addr >= 0xFE00 && addr <= 0xFE9F {
			if p.state().canAccessOAM() {
				return p.oam.Read(addr)
			}
			return 0xFF
		}
	}
	// Todo: Map IO Registers
	return 0xFF
}

func (p *PPU) Write(addr uint16, val byte) {
	switch addr {
	case consts.AddrVBK:
		p.vramHi = val&1 != 0
	case consts.AddrLCDC:
		oldEnabled := p.lcdEnabled()
		p.lcdc = val
		if newEnabled := p.lcdEnabled(); oldEnabled != newEnabled {
			if !newEnabled {
				p.screenOut <- emptyScreen
			} else {
				p.setLy(0)
				p.phaseIdx = 0
				p.phases[0].start(p)
				p.curScreen = newScreen()
			}
		}
	case consts.AddrSTAT:
		p.ie = lcdInterrupts(val) & liALL
	case consts.AddrSCROLLY:
		p.scrollY = val
	case consts.AddrSCROLLX:
		p.scrollX = val
	case consts.AddrLY:
		// readOnly...
		return
	case consts.AddrLYC:
		p.lyc = val
	case consts.AddrBGP:
		p.bgPal = gbPalette(val)
	case consts.AddrOBJECTPALETTE0:
		p.objPal = (p.objPal & 0xFF00) | gbPalette(val)
	case consts.AddrOBJECTPALETTE1:
		p.objPal = (p.objPal & 0x00FF) | gbPalette(val)
	case consts.AddrWY:
		p.winY = val
	case consts.AddrWX:
		p.winX = val
	default:
		if addr >= 0x8000 && addr <= 0x9FFF {
			if p.state().canAccessVRAM() {
				if p.mmu.GBC() && p.vramHi {
					p.vram1.Write(addr, val)
					return
				}
				p.vram0.Write(addr, val)
			}
		} else if addr >= 0xFE00 && addr <= 0xFE9F {
			if p.state().canAccessOAM() {
				p.oam.Write(addr, val)
			}
		}
	}
}

func (p *PPU) useWndAndBg() bool {
	return p.lcdc&0x01 != 0
}

func (p *PPU) useObjects() bool {
	return p.lcdc&0x02 != 0
}

func (p *PPU) spriteHeight() byte {
	if p.lcdc&0x04 != 0 {
		return 16
	}
	return 8
}

func (p *PPU) bgTileDisplayAddr() uint16 {
	if p.lcdc&0x08 == 0 {
		return 0x9800
	}
	return 0x9c00
}

func (p *PPU) bgTileDataAddr() uint16 {
	if p.lcdc&0x10 == 0 {
		return 0x9000
	}
	return 0x8000
}
func (p *PPU) bgWndTileDataSigned() bool {
	return (p.lcdc & 0x10) == 0
}

func (p *PPU) useWnd() bool {
	return p.lcdc&0x20 != 0
}

func (p *PPU) wndTileMapDisplayAddr() uint16 {
	if p.lcdc&0x40 == 0 {
		return 0x9800
	}
	return 0x9c00
}

func (p *PPU) lcdEnabled() bool {
	return p.lcdc&0x80 != 0
}

// Step the PPU for one M-Cycle
func (p *PPU) Step() {
	// ppu runs at 4 times the speed of the cpu
	for i := 0; i < 4; i++ {
		if p.lcdEnabled() {
			p.stepOne()
		}
		if dma := p.dma; dma != nil {
			dma.Step(p)
		}
	}
}

func (p *PPU) requstLcdcInterrupt(i lcdInterrupts) {
	if (p.ie & i) == i {
		p.mmu.RequestInterrupt(mmu.IRQLCDStat)
	}
}

func (p *PPU) setLy(v byte) {
	p.ly = v
	p.ticksInLine = 0
	if p.ly == p.lyc {
		p.requstLcdcInterrupt(liCoincidence)
	}
}

func (p *PPU) stepOne() {
	p.ticksInLine++
	if !p.phases[p.phaseIdx].step(p) {
		if p.ticksInLine == 4 && p.state() == sVBlank && p.ly == 153 {
			p.setLy(0)
		}
	} else {
		incLy := p.ly != 0 || p.state() != sVBlank
		for {
			if p.phaseIdx++; p.phaseIdx == len(p.phases) {
				if incLy {
					p.setLy(p.ly + 1)
				}
				p.phaseIdx = 0
			}
			if p.phases[p.phaseIdx].start(p) {
				break
			}
		}
	}

}
