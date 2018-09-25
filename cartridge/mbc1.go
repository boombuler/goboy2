package cartridge

type mbc1 struct {
	battery   Battery
	rombanks  []rombank
	activerom int

	rambanks   []rambank
	activeram  int
	ramMode    bool
	ramEnabled bool
}

func createMBC1(c *Cartridge, data []byte, bat Battery) (MBC, error) {
	m := new(mbc1)
	m.battery = bat
	for rs := c.ROMSize; rs > 0; rs -= rombankSize {
		m.rombanks = append(m.rombanks, rombank(data[c.ROMSize-rs:c.ROMSize-rs+rombankSize]))
	}
	m.activerom = 1
	m.rambanks = make([]rambank, c.RAMSize/rambankSize)
	for _, rb := range m.rambanks {
		for i := 0; i < rambankSize; i++ {
			rb[i] = 0xFF
		}
	}

	m.activeram = 0
	m.ramEnabled = false
	m.ramMode = false
	m.loadRAM()
	return m, nil
}

func (m *mbc1) hasRam() bool {
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
		if m.ramEnabled && m.hasRam() {
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
		m.activerom = (m.activerom & 0x60) | int(value&0x1F)
	case addr >= 0x4000 && addr <= 0x5FFF:
		m.activeram = int(value & 0x03)
		if !m.ramMode {
			// ROM mode: Set high bits of bank
			m.activerom = (m.activerom & 0x1F) | ((int(value) & 3) << 5)
		}
	case addr >= 0x6000 && addr <= 0x7FFF:
		m.ramMode = value&0x01 != 0x00
	case addr >= 0xA000 && addr <= 0xBFFF:
		if m.ramEnabled && m.hasRam() {
			m.rambanks[m.getRAMBank()][addr-0xA000] = value
		}
	}
}

func (m *mbc1) getHiROMBank() int {
	bank := m.activerom
	if bank%0x20 == 0 {
		bank++
	}
	if m.ramMode {
		bank &= 0x1F
		bank |= (m.activeram << 5)
	}
	return bank % len(m.rombanks)
}
func (m *mbc1) getLoROMBank() int {
	if !m.ramMode {
		return 0
	}
	return (m.activeram << 5) % len(m.rombanks)
}

func (m *mbc1) getRAMBank() int {
	if m.ramMode {
		return m.activeram % len(m.rambanks)
	}
	return 0
}

func (m *mbc1) Shutdown() {
	m.saveRAM()
}

func (m *mbc1) saveRAM() {
	if m.battery != nil && m.hasRam() {
		w := m.battery.Open()
		for i := 0; i < len(m.rambanks); i++ {
			w.Write(m.rambanks[i][:])
		}
		w.Close()
	}
}
func (m *mbc1) loadRAM() {
	if m.battery != nil && m.battery.HasData() && m.hasRam() {
		r := m.battery.Open()
		for i := 0; i < len(m.rambanks); i++ {
			r.Read(m.rambanks[i][:])
		}
		r.Close()
	}
}
