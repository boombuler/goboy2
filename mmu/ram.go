package mmu

type workingRAM struct {
	bank0 [4096]byte
	bank1 [4096]byte
}

func (wr *workingRAM) Read(addr uint16) byte {
	if addr >= 0xE000 {
		// shadow ram...
		addr -= 0x2000
	}
	switch {
	case addr >= 0xC000 && addr <= 0xCFFF:
		return wr.bank0[addr-0xC000]
	case addr >= 0xD000 && addr <= 0xDFFF:
		return wr.bank1[addr-0xD000]
	}
	return 0xFF
}

func (wr *workingRAM) Write(addr uint16, val byte) {
	if addr >= 0xE000 {
		// shadow ram...
		addr -= 0x2000
	}
	switch {
	case addr >= 0xC000 && addr <= 0xCFFF:
		wr.bank0[addr-0xC000] = val
	case addr >= 0xD000 && addr <= 0xDFFF:
		wr.bank1[addr-0xD000] = val
	}
}
