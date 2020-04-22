package cartridge

import (
	"io"
)

type mbc5 struct {
	battery   BatteryStream
	rombanks  []rombank
	activerom int

	rambanks   []rambank
	activeram  int
	ramEnabled bool
}

func createMBC5(c *Cartridge, data []byte, bat Battery) (MBC, error) {
	m := new(mbc5)
	if bat != nil {
		m.battery = bat.Open()
	}
	for rs := c.ROMSize; rs > 0; rs -= rombankSize {
		m.rombanks = append(m.rombanks, rombank(data[c.ROMSize-rs:c.ROMSize-rs+rombankSize]))
	}
	m.activerom = 1
	ramBankCnt := c.RAMSize / rambankSize
	if ramBankCnt == 0 {
		ramBankCnt = 1
	}
	m.rambanks = make([]rambank, ramBankCnt)
	for _, rb := range m.rambanks {
		for i := 0; i < rambankSize; i++ {
			rb[i] = 0xFF
		}
	}

	m.activeram = 0
	m.loadRAM()
	return m, nil
}

func (m *mbc5) hasRam() bool {
	return len(m.rambanks) > 0
}

func (m *mbc5) Read(addr uint16) byte {
	if addr < rombankSize {
		return m.rombanks[0].Read(addr)
	}
	if addr >= rombankSize && addr < 2*rombankSize {
		return m.rombanks[m.activerom%len(m.rombanks)].Read(addr)
	}
	if addr >= 0xA000 && addr < 0xC000 {
		if m.ramEnabled && m.hasRam() {
			return m.rambanks[m.activeram%len(m.rambanks)][addr-0xA000]
		}
	}

	return 0xFF
}

func (m *mbc5) Write(addr uint16, value byte) {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		m.ramEnabled = value&0x0F == 0x0A
		if !m.ramEnabled {
			m.saveRAM()
		}
	case addr >= 0x2000 && addr < 0x3000:
		m.activerom = (m.activerom & 0x100) | int(value)
	case addr >= 0x3000 && addr < 0x4000:
		m.activerom = (m.activerom & 0x0ff) | (int(value&1) << 8)
	case addr >= 0x4000 && addr < 0x6000:
		bank := int(value & 0x0f)
		if bank < len(m.rambanks) {
			m.activeram = bank
		}
	case addr >= 0xA000 && addr < 0xC000 && m.ramEnabled:
		m.rambanks[m.activeram%len(m.rambanks)][addr-0xA000] = value
	}
}

func (m *mbc5) Shutdown() {
	m.saveRAM()
	m.battery.Close()
	m.battery = nil
}

func (m *mbc5) saveRAM() {
	if m.battery != nil && m.hasRam() {
		m.battery.Seek(0, io.SeekStart)
		for i := 0; i < len(m.rambanks); i++ {
			m.battery.Write(m.rambanks[i][:])
		}
	}
}

func (m *mbc5) loadRAM() {
	if m.battery != nil && m.hasRam() {
		m.battery.Seek(0, io.SeekStart)
		for i := 0; i < len(m.rambanks); i++ {
			m.battery.Read(m.rambanks[i][:])
		}
	}
}
