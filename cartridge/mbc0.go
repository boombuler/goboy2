package cartridge

import (
	"fmt"
)

type mbc0 struct {
	rom     []rombank
	ram     rambank
	hasRam  bool
	battery Battery
}

func createMBC0(c *Cartridge, data []byte, hasRam bool, bat Battery) (MBC, error) {
	if len(data) < 0x8000 {
		return nil, fmt.Errorf("Invalid ROM size: %v", len(data))
	}
	m := &mbc0{
		rom: []rombank{
			data[:rombankSize],
			data[rombankSize : rombankSize*2],
		},
		hasRam:  hasRam,
		battery: bat,
	}
	m.loadRAM()
	return m, nil
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

func (m *mbc0) saveRAM() {
	if m.battery != nil && m.hasRam {
		f := m.battery.Open()
		f.Write(m.ram[:])
		f.Close()
	}
}

func (m *mbc0) Shutdown() {
	m.saveRAM()
}

func (m *mbc0) loadRAM() {
	if m.battery != nil && m.hasRam && m.battery.HasData() {
		r := m.battery.Open()
		r.Read(m.ram[:])
		r.Close()
	}
}
