package cartridge

import (
	"io"
)

type mbc3 struct {
	battery    bool
	rombanks   []rombank
	activerom  int
	rambanks   []rambank
	activeram  int
	ramEnabled bool
	rtc        *rtc
}

func createMBC3(c *Cartridge, data []byte, hasTimer, hasBattery bool) (MBC, error) {
	m := new(mbc3)
	m.battery = hasBattery
	for rs := c.ROMSize; rs > 0; rs -= rombankSize {
		m.rombanks = append(m.rombanks, rombank(data[c.ROMSize-rs:c.ROMSize-rs+rombankSize]))
	}
	m.activerom = 1
	m.rambanks = make([]rambank, c.RAMSize/rambankSize)
	m.activeram = 0
	m.ramEnabled = false
	if hasTimer {
		m.rtc = newRealTimeClock()
	}
	return m, nil
}

func (m *mbc3) hasRam() bool {
	return len(m.rambanks) > 0
}

func (m *mbc3) Read(addr uint16) byte {
	if addr < rombankSize {
		return m.rombanks[0].Read(addr)
	}
	if addr >= rombankSize && addr < 2*rombankSize {
		return m.rombanks[m.activerom].Read(addr)
	}
	if addr >= 0xA000 && addr < 0xC000 {
		if m.ramEnabled {
			if m.activeram <= 0x03 && m.hasRam() {
				return m.rambanks[m.activeram][addr-0xA000]
			} else if m.rtc != nil {
				return m.rtc.Read(m.activeram)
			}
		}
	}
	return 0x00
}

func (m *mbc3) Write(addr uint16, value byte) {
	if addr >= 0xA000 && addr < 0xC000 {
		if m.ramEnabled {
			if m.activeram <= 0x03 && m.hasRam() {
				m.rambanks[m.activeram][addr-0xA000] = value
			} else if m.rtc != nil {
				m.rtc.Write(m.activeram, value)
			}
		}
	} else if addr >= 0x0000 && addr <= 0x1FFF {
		m.ramEnabled = value&0x0F == 0x0A
	} else if addr >= 0x2000 && addr <= 0x3FFF {
		m.activerom = int(value & 0x7F)
		if m.activerom == 0x00 {
			m.activerom = 0x01
		}
	} else if addr >= 0x4000 && addr <= 0x5FFF {
		if value <= 0x0C {
			m.activeram = int(value)
		}
	} else if addr >= 0x6000 && addr <= 0x7FFF {
		if m.rtc != nil {
			m.rtc.WriteLatch(value)
		}
	}
}

func (m *mbc3) DumpRAM(w io.Writer) {
	if m.battery {
		if m.hasRam() {
			for _, rb := range m.rambanks {
				w.Write(rb[:])
			}
		}
		if m.rtc != nil {
			m.rtc.Dump(w)
		}
	}
}
func (m *mbc3) LoadRAM(r io.Reader) {
	if m.battery {
		if m.hasRam() {
			for _, rb := range m.rambanks {
				r.Read(rb[:])
			}
		}
		if m.rtc != nil {
			m.rtc.Load(r)
		}
	}
}
