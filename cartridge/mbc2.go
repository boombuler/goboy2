package cartridge

import (
	"fmt"
)

type mbc2 struct {
	rombanks   []rombank
	rambank    [0x0200]byte
	activerom  int
	ramenabled bool
	battery    Battery
}

func createMBC2(c *Cartridge, data []byte, bat Battery) (MBC, error) {
	if len(data) < 0x8000 {
		return nil, fmt.Errorf("Invalid ROM size: %v", len(data))
	}
	m := new(mbc2)
	for rs := c.ROMSize; rs > 0; rs -= rombankSize {
		m.rombanks = append(m.rombanks, rombank(data[c.ROMSize-rs:c.ROMSize-rs+rombankSize]))
	}
	m.activerom = 1
	m.ramenabled = false
	m.battery = bat
	m.loadRAM()
	return m, nil
}

func (m *mbc2) Read(addr uint16) byte {
	if addr < rombankSize {
		return m.rombanks[0].Read(addr)
	}
	if addr >= rombankSize && addr < 2*rombankSize {
		return m.rombanks[m.activerom].Read(addr)
	}
	if addr >= 0xA000 && addr < 0xA200 {
		if m.ramenabled {
			return m.rambank[addr-0xA000] & 0x0F
		}
	}

	return 0x00
}
func (m *mbc2) Write(addr uint16, value byte) {
	if addr <= 0x1FFF && (addr&0x0100 == 0x0000) {
		m.ramenabled = value&0x0F == 0x0A
		if !m.ramenabled {
			m.saveRAM()
		}
	}
	if addr >= 0x2000 && addr <= 0x3FFF && (addr&0x0100) == 0x0100 {
		m.activerom = int(value & 0x0F)
		if m.activerom == 0 || m.activerom >= len(m.rombanks) {
			m.activerom = 1
		}
	}
	if m.ramenabled && addr >= 0xA000 && addr < 0xA200 {
		m.rambank[addr-0xA000] = value & 0x0F
	}
}

func (m *mbc2) Shutdown() {
	m.saveRAM()
}
func (m *mbc2) saveRAM() {
	if m.battery != nil {
		w := m.battery.Open()
		w.Write(m.rambank[:])
		w.Close()
	}
}
func (m *mbc2) loadRAM() {
	if m.battery != nil && m.battery.HasData() {
		r := m.battery.Open()
		r.Read(m.rambank[:])
		r.Close()
	}
}
