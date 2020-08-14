package mmu

import "github.com/boombuler/goboy2/consts"

type gbcRegisters struct {
	mmu   MMU
	reg6C byte
	reg72 byte
	reg73 byte
	reg74 byte
	reg75 byte
}

func newGBCRegisters(mmu MMU) *gbcRegisters {
	return &gbcRegisters{
		mmu:   mmu,
		reg6C: 0x00,
		reg72: 0x00,
		reg73: 0x00,
		reg74: 0x00,
		reg75: 0x00,
	}
}

func (r *gbcRegisters) IOAddrs() []uint16 {
	return []uint16{
		0xFF6C, 0xFF72, 0xFF73, 0xFF74, 0xFF75,
	}
}

func (r *gbcRegisters) Read(addr uint16) byte {
	switch addr {
	case 0xFF6C:
		// 1 bit register
		if r.mmu.EmuMode() == consts.DMG {
			return 0xFF
		}
		return r.reg6C
	case 0xFF72:
		return r.reg72
	case 0xFF73:
		return r.reg73
	case 0xFF74:
		if r.mmu.EmuMode() == consts.DMG {
			return 0xFF
		}
		return r.reg74
	case 0xFF75:
		return r.reg75 | 0x8F

	}
	return 0xFF
}

func (r *gbcRegisters) Write(addr uint16, val byte) {
	switch addr {
	case 0xFF6C:
		// 1 bit register
		r.reg6C = val & 0xFE
	case 0xFF72:
		r.reg72 = val
	case 0xFF73:
		r.reg73 = val
	case 0xFF74:
		r.reg74 = val
	case 0xFF75:
		r.reg75 = val
	}
}
