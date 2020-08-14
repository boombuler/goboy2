package mmu

import (
	"github.com/boombuler/goboy2/consts"
)

type rambank [4096]byte
type workingRAM struct {
	mmu          MMU
	selectedBank int
	banks        []rambank
}

func newWorkingRAM(mmu MMU) IODevice {
	wr := new(workingRAM)
	wr.mmu = mmu

	if mmu.HardwareCompat() == consts.GBC {
		wr.banks = make([]rambank, 8)
		mmu.AddIODevice(wr, consts.AddrSVBK)
	} else {
		wr.banks = make([]rambank, 2)
	}
	wr.selectedBank = 1
	return wr
}

func (wr *workingRAM) Read(addr uint16) byte {
	if addr == consts.AddrSVBK {
		if wr.mmu.EmuMode() == consts.GBC {
			return byte(wr.selectedBank) | 0xF8
		}
		return 0xFF
	}
	if addr >= 0xE000 {
		// shadow ram...
		addr -= 0x2000
	}
	switch {
	case addr >= 0xC000 && addr <= 0xCFFF:
		return wr.banks[0][addr-0xC000]
	case addr >= 0xD000 && addr <= 0xDFFF:
		return wr.banks[wr.selectedBank][addr-0xD000]
	}
	return 0xFF
}

func (wr *workingRAM) Write(addr uint16, val byte) {
	if addr == consts.AddrSVBK {
		if wr.mmu.EmuMode() == consts.GBC {
			wr.selectedBank = int(val & 0x07)
			if wr.selectedBank == 0 {
				wr.selectedBank = 1
			}
		}
		return
	}

	if addr >= 0xE000 {
		// shadow ram...
		addr -= 0x2000
	}
	switch {
	case addr >= 0xC000 && addr <= 0xCFFF:
		wr.banks[0][addr-0xC000] = val
	case addr >= 0xD000 && addr <= 0xDFFF:
		wr.banks[wr.selectedBank][addr-0xD000] = val
	}
}
