package cartridge

import (
	"io"
)

type mbc1 struct {
	battery   bool
	rombanks  []rombank
	activerom int

	rambanks   []rambank
	activeram  int
	ramMode    bool
	ramEnabled bool
}

func createMBC1(c *Cartridge, data []byte, hasBattery bool) (MBC, error) {
	m := new(mbc1)
	m.battery = hasBattery
	for rs := c.ROMSize; rs > 0; rs -= rombankSize {
		m.rombanks = append(m.rombanks, rombank(data[c.ROMSize-rs:c.ROMSize-rs+rombankSize]))
	}
	m.activerom = 1
	m.rambanks = make([]rambank, c.RAMSize/rambankSize)
	m.activeram = 0
	m.ramEnabled = c.RAMSize > 0
	m.ramMode = false
	return m, nil
}

func (m *mbc1) hasRam() bool {
	return len(m.rambanks) > 0
}

func (m *mbc1) Read(addr uint16) byte {
	if addr < rombankSize {
		return m.rombanks[0].Read(addr)
	}
	if addr >= rombankSize && addr < 2*rombankSize {
		return m.rombanks[m.activerom].Read(addr)
	}
	if addr >= 0xA000 && addr < 0xC000 {
		if m.ramEnabled && m.hasRam() {
			if m.ramMode {
				return m.rambanks[m.activeram][addr-0xA000]
			}
			return m.rambanks[0][addr-0xA000]
		}
	}

	return 0x00
}

func (m *mbc1) Write(addr uint16, value byte) {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		if m.ramMode && m.hasRam() {
			m.ramEnabled = value&0x0F == 0x0A
		}
	case addr >= 0x2000 && addr <= 0x3FFF:
		// Set lower 5 bits of ROM bank (skipping #0)
		value &= 0x1F
		if value == 0x00 {
			value = 0x01
		}
		m.activerom = (m.activerom & 0x60) | int(value)
	case addr >= 0x4000 && addr <= 0x5FFF:
		if m.ramMode {
			m.activeram = int(value & 0x03)
		} else {
			// ROM mode: Set high bits of bank
			m.activerom = (m.activerom & 0x1F) | ((int(value) & 3) << 5)
		}
	case addr >= 0x6000 && addr <= 0x7FFF:
		m.ramMode = value&0x01 != 0x00
	case addr >= 0xA000 && addr <= 0xBFFF:
		if m.ramEnabled && m.hasRam() {
			if m.ramMode {
				m.rambanks[m.activeram][addr-0xA000] = value
			} else {
				m.rambanks[0][addr-0xA000] = value
			}
		}
	}
}
func (m *mbc1) DumpRAM(w io.Writer) {
	if m.battery && m.hasRam() {
		for _, rb := range m.rambanks {
			w.Write(rb[:])
		}
	}
}
func (m *mbc1) LoadRAM(r io.Reader) {
	if m.battery && m.hasRam() {
		for _, rb := range m.rambanks {
			r.Read(rb[:])
		}
	}
}
