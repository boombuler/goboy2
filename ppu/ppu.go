package ppu

import (
	"goboy2/mmu"
	"image"
)

const (
	addrLCDC           uint16 = 0xFF40
	addrSTAT           uint16 = 0xFF41
	addrSCROLLY        uint16 = 0xFF42
	addrSCROLLX        uint16 = 0xFF43
	addrLY             uint16 = 0xFF44
	addrLYC            uint16 = 0xFF45
	addrBGP            uint16 = 0xFF47
	addrOBJECTPALETTE0 uint16 = 0xFF48
	addrOBJECTPALETTE1 uint16 = 0xFF49
	addrWY             uint16 = 0xFF4A
	addrWX             uint16 = 0xFF4B
)

type PPU struct {
	mmu            mmu.MMU
	phaseIdx       int
	phases         []ppuPhase
	visibleSprites []*spriteData
	ticksInLine    uint16

	lcdc      byte
	ie        lcdInterrupts
	ly        byte
	lyc       byte
	scrollY   byte
	scrollX   byte
	winX      byte
	winY      byte
	bgPal     palette
	obj0      palette
	obj1      palette
	screenOut chan<- *image.RGBA
	curScreen *image.RGBA

	vram vRAM
	oam  oam
}

// DisplayWidth is the width of the output pictures
const DisplayWidth int = 160

// DisplayHeight is the height of the output pictures
const DisplayHeight int = 144

// New creates a new ppu and connects it to the given mmu
func New(mmu mmu.MMU, screen chan<- *image.RGBA) *PPU {
	ppu := &PPU{
		mmu:       mmu,
		vram:      newVRAM(),
		oam:       newOAM(),
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
	mmu.AddIODevice(ppu, addrLCDC, addrSTAT, addrSCROLLY, addrSCROLLX, addrLY,
		addrLYC, addrBGP, addrOBJECTPALETTE0, addrOBJECTPALETTE1, addrWY, addrWX)
	return ppu
}

func (p *PPU) state() ppuState {
	return p.phases[p.phaseIdx].state()
}

func (p *PPU) Read(addr uint16) byte {
	switch addr {
	case addrLCDC:
		return p.lcdc
	case addrSTAT:
		var coincidence byte
		if p.ly == p.lyc {
			coincidence = 1
		}
		return byte(p.ie) | (coincidence << 2) | (byte(p.state()) & 0x03)
	case addrSCROLLY:
		return p.scrollY
	case addrSCROLLX:
		return p.scrollX
	case addrLY:
		return p.ly
	case addrLYC:
		return p.lyc
	case addrBGP:
		return byte(p.bgPal)
	case addrOBJECTPALETTE0:
		return byte(p.obj0)
	case addrOBJECTPALETTE1:
		return byte(p.obj1)
	case addrWY:
		return p.winY
	case addrWX:
		return p.winX
	default:
		if addr >= 0x8000 && addr <= 0x9FFF {
			if p.state().canAccessVRAM() {
				return p.vram.Read(addr)
			}
			return 0xFF
		} else if addr >= 0xFE00 && addr <= 0xFE9F {
			if p.state().canAccessVRAM() {
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
	case addrLCDC:
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
	case addrSTAT:
		p.ie = lcdInterrupts(val) & liALL
	case addrSCROLLY:
		p.scrollY = val
	case addrSCROLLX:
		p.scrollX = val
	case addrLY:
		// readOnly...
		return
	case addrLYC:
		p.lyc = val
	case addrBGP:
		p.bgPal = palette(val)
	case addrOBJECTPALETTE0:
		p.obj0 = palette(val)
	case addrOBJECTPALETTE1:
		p.obj1 = palette(val)
	case addrWY:
		p.winY = val
	case addrWX:
		p.winX = val
	default:
		if addr >= 0x8000 && addr <= 0x9FFF {
			if p.state().canAccessVRAM() {
				p.vram.Write(addr, val)
			}
		} else if addr >= 0xFE00 && addr <= 0xFE9F {
			if p.state().canAccessVRAM() {
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
	if !p.lcdEnabled() {
		return
	}

	// ppu runs at 4 times the speed of the cpu
	for i := 0; i < 4; i++ {
		p.stepOne()
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
		for {
			if p.phaseIdx++; p.phaseIdx == len(p.phases) {
				p.setLy(p.ly + 1)
				p.phaseIdx = 0
			}
			if p.phases[p.phaseIdx].start(p) {
				break
			}
		}
	}

}
