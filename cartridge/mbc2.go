package cartridge

import (
	"fmt"
	"io"
)

type mbc2 struct {
	rombanks   []rombank
	rambank    [0x0200]byte
	activerom  int
	ramenabled bool
	battery    bool
}

func createMBC2(c *Cartridge, data []byte, hasBat bool) (MBC, error) {
	if len(data) < 0x8000 {
		return nil, fmt.Errorf("Invalid ROM size: %v", len(data))
	}
	m := new(mbc2)
	for rs := c.ROMSize; rs > 0; rs -= rombankSize {
		m.rombanks = append(m.rombanks, rombank(data[c.ROMSize-rs:c.ROMSize-rs+rombankSize]))
	}
	m.activerom = 1
	m.ramenabled = false
	m.battery = hasBat
	return m, nil
}

func (m *mbc2) HasBattery() bool {
	return m.battery
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
func (m *mbc2) DumpRAM(w io.Writer) {
	if m.battery {
		w.Write(m.rambank[:])
	}
}
func (m *mbc2) LoadRAM(r io.Reader) {
	if m.battery {
		r.Read(m.rambank[:])
	}
}
