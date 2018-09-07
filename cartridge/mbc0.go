package cartridge

import (
	"fmt"
	"io"
)

type mbc0 struct {
	rom             []rombank
	ram             rambank
	hasRam, battery bool
}

func createMBC0(c *Cartridge, data []byte, hasRam, hasBat bool) (MBC, error) {
	if len(data) < 0x8000 {
		return nil, fmt.Errorf("Invalid ROM size: %v", len(data))
	}
	return &mbc0{
		rom: []rombank{
			data[:rombankSize],
			data[rombankSize : rombankSize*2],
		},
		hasRam:  hasRam,
		battery: hasBat,
	}, nil
}

func (m *mbc0) Read(addr uint16) byte {
	if addr < rombankSize*2 {
		return m.rom[addr/rombankSize].Read(addr)
	} else if m.hasRam && addr >= 0xA000 && addr <= 0xBFFF {
		return m.ram[addr-0xA000]
	}
	return 0x00
}
func (m *mbc0) Write(addr uint16, value byte) {
	if m.hasRam && addr >= 0xA000 && addr <= 0xBFFF {
		m.ram[addr-0xA000] = value
	}
}
func (m *mbc0) DumpRAM(w io.Writer) {
	if m.battery && m.hasRam {
		w.Write(m.ram[:])
	}
}
func (m *mbc0) LoadRAM(r io.Reader) {
	if m.battery && m.hasRam {
		r.Read(m.ram[:])
	}
}
