package cartridge

import (
	"fmt"
)

const mbc2RAMSize = 0x0200

type mbc2 struct {
	rombanks []rombank
	rambank  [mbc2RAMSize]byte
	romb     int
	ramg     bool
	battery  Battery
}

func createMBC2(c *Cartridge, data []byte, bat Battery) (MBC, error) {
	if len(data) < 0x8000 {
		return nil, fmt.Errorf("Invalid ROM size: %v", len(data))
	}
	m := new(mbc2)
	for rs := c.ROMSize; rs > 0; rs -= rombankSize {
		m.rombanks = append(m.rombanks, rombank(data[c.ROMSize-rs:c.ROMSize-rs+rombankSize]))
	}
	m.romb = 1
	m.ramg = false
	m.battery = bat
	m.loadRAM()
	return m, nil
}

func (m *mbc2) Read(addr uint16) byte {
	if addr < rombankSize {
		return m.rombanks[0].Read(addr)
	}
	if addr >= rombankSize && addr < 2*rombankSize {

		return m.rombanks[m.romb].Read(addr)
	}
	if addr >= 0xA000 && addr < 0xC000 {
		if m.ramg {
			return m.rambank[(addr-0xA000)%mbc2RAMSize] | 0xF0
		}
	}

	return 0xFF
}
func (m *mbc2) Write(addr uint16, value byte) {
	if addr < 0x4000 {
		if addr&0x0100 == 0 {
			m.ramg = value&0x0F == 0x0A
			if !m.ramg {
				m.saveRAM()
			}
		} else {
			val := int(value & 0x0F)
			if val == 0 {
				val = 1
			}
			m.romb = val % len(m.rombanks)
		}
	} else if m.ramg && addr >= 0xA000 && addr < 0xC000 {
		m.rambank[(addr-0xA000)%mbc2RAMSize] = value
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
