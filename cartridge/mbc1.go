package cartridge

import "bytes"

type mbc1 struct {
	battery  Battery
	rombanks []rombank
	rambanks []rambank

	multicard  bool
	bank1      int
	bank2      int
	ramEnabled bool
	mode       byte
}

const (
	logoStart = 0x0104
	logoEnd   = 0x0134
)

func createMBC1(c *Cartridge, data []byte, bat Battery) (MBC, error) {
	m := new(mbc1)
	m.battery = bat
	for rs := c.ROMSize; rs > 0; rs -= rombankSize {
		m.rombanks = append(m.rombanks, rombank(data[c.ROMSize-rs:c.ROMSize-rs+rombankSize]))
	}
	m.rambanks = make([]rambank, c.RAMSize/rambankSize)
	for _, rb := range m.rambanks {
		for i := 0; i < rambankSize; i++ {
			rb[i] = 0xFF
		}
	}

	m.multicard = false
	if len(m.rombanks) > 32 {
		// try to detect multicard mbc1 cards:

		m.multicard = true
		logoData := m.rombanks[0][logoStart:logoEnd]

		for i := 1; i < 4; i++ {
			bankNo := (i << 5)
			if len(m.rombanks) > bankNo {
				if bytes.Compare(logoData, m.rombanks[bankNo][logoStart:logoEnd]) != 0 {
					m.multicard = false
					break
				}
			} else {
				break
			}
		}
	}

	m.bank1 = 1
	m.bank2 = 0
	m.mode = 0
	m.loadRAM()
	return m, nil
}

func (m *mbc1) hasRAM() bool {
	return len(m.rambanks) > 0
}

func (m *mbc1) Read(addr uint16) byte {
	if addr < rombankSize {
		return m.rombanks[m.getLoROMBank()].Read(addr)
	}
	if addr >= rombankSize && addr < 2*rombankSize {
		return m.rombanks[m.getHiROMBank()].Read(addr)
	}
	if addr >= 0xA000 && addr < 0xC000 {
		if m.ramEnabled && m.hasRAM() {
			return m.rambanks[m.getRAMBank()][addr-0xA000]
		}
	}

	return 0xFF
}

func (m *mbc1) Write(addr uint16, value byte) {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		m.ramEnabled = value&0x0F == 0x0A
		if !m.ramEnabled {
			m.saveRAM()
		}
	case addr >= 0x2000 && addr <= 0x3FFF:
		val := int(value & 0x1F)
		if val == 0 {
			val = 1
		}
		m.bank1 = val
	case addr >= 0x4000 && addr <= 0x5FFF:
		m.bank2 = int(value & 0x03)
	case addr >= 0x6000 && addr <= 0x7FFF:
		m.mode = value & 0x01
	case addr >= 0xA000 && addr <= 0xBFFF:
		if m.ramEnabled && m.hasRAM() {
			m.rambanks[m.getRAMBank()][addr-0xA000] = value
		}
	}
}

func (m *mbc1) getBank2Shift() byte {
	if m.multicard {
		return 4
	}
	return 5
}

func (m *mbc1) getHiROMBank() int {
	bank1 := m.bank1
	if m.multicard {
		bank1 = bank1 & 0x0F
	}

	return ((m.bank2 << m.getBank2Shift()) | bank1) % len(m.rombanks)
}

func (m *mbc1) getLoROMBank() int {
	if m.mode == 0x00 {
		return 0
	}
	return (m.bank2 << m.getBank2Shift()) % len(m.rombanks)
}

func (m *mbc1) getRAMBank() int {
	if m.mode == 0x01 {
		return m.bank2 % len(m.rambanks)
	}
	return 0
}

func (m *mbc1) Shutdown() {
	m.saveRAM()
}

func (m *mbc1) saveRAM() {
	if m.battery != nil && m.hasRAM() {
		w := m.battery.Open()
		for i := 0; i < len(m.rambanks); i++ {
			w.Write(m.rambanks[i][:])
		}
		w.Close()
	}
}

func (m *mbc1) loadRAM() {
	if m.battery != nil && m.battery.HasData() && m.hasRAM() {
		r := m.battery.Open()
		for i := 0; i < len(m.rambanks); i++ {
			r.Read(m.rambanks[i][:])
		}
		r.Close()
	}
}
