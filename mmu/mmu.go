package mmu

type MMU interface {
	IODevice
	RequestInterrupt(i IRQ)
	GetCurrentIterrupt() IRQ
	ConnectPPU(ppu IODevice)
	LoadCartridge(cartridge IODevice)
	AddIODevice(d IODevice, addrs ...uint16)
	Step()
}

type mmuImpl struct {
	ioDevices []IODevice
	cartridge IODevice
	ppu       IODevice
	ram       [2 * 4096]byte
	zpram     [127]byte
	dma       *dmaTransfer
}

type IODevice interface {
	Read(addr uint16) byte
	Write(addr uint16, value byte)
}

const AddrBootmodeFlag = 0xFF50

type bootMode byte

func (bm *bootMode) Read(addr uint16) byte {
	if byte(*bm) == 0x00 {
		return 0x00
	}
	return 0xFF
}
func (bm *bootMode) Write(addr uint16, value byte) {
	if *bm == 0x00 {
		*bm = bootMode(value)
	}
}

func New() MMU {
	res := &mmuImpl{
		ioDevices: make([]IODevice, 256),
	}
	res.dma = &dmaTransfer{mmu: res}
	res.AddIODevice(new(irqHandler), AddrIRQFlags, AddrIRQEnabled)
	res.AddIODevice(new(bootMode), AddrBootmodeFlag)
	res.AddIODevice(res.dma, AddrDMATransfer)
	return res
}

func (m *mmuImpl) Step() {
	m.dma.step()
}

func (m *mmuImpl) AddIODevice(d IODevice, addrs ...uint16) {
	for _, address := range addrs {
		m.ioDevices[address&0xFF] = d
	}
}

func (m *mmuImpl) LoadCartridge(cartridge IODevice) {
	m.cartridge = cartridge
}

func (m *mmuImpl) ConnectPPU(ppu IODevice) {
	m.ppu = ppu
}

func (m *mmuImpl) Read(addr uint16) byte {
	// [FF80-FFFE] Zero-page RAM
	if addr >= 0xFF80 && addr < 0xFFFF {
		return m.zpram[addr-0xFF80]
	} else if m.dma.blockMemoryAccess(addr) {
		return 0xFF
	}

	switch {
	// [0000-3FFF] Cartridge ROM, bank 0
	case addr >= 0x0000 && addr <= 0x3FFF:
		if addr <= 0x00FF && m.Read(AddrBootmodeFlag) == 0x00 {
			return BOOTROM[addr]
		}
		return m.cartridge.Read(addr)
	// [4000-7FFF] Cartridge ROM, other banks
	case addr >= 0x4000 && addr <= 0x7FFF:
		return m.cartridge.Read(addr)
	// [8000-9FFF] Graphics RAM
	case addr >= 0x8000 && addr <= 0x9FFF:
		return m.ppu.Read(addr)
	// [A000-BFFF] Cartridge (External) RAM
	case addr >= 0xA000 && addr <= 0xBFFF:
		return m.cartridge.Read(addr)
	// [C000-DFFF] Working RAM
	case addr >= 0xC000 && addr <= 0xDFFF:
		return m.ram[addr-0xC000]
	// [E000-FDFF] Working RAM (shadow)
	case addr >= 0xE000 && addr <= 0xFDFF:
		return m.ram[addr-0xE000]
	// [FE00-FE9F] Graphics: sprite information
	case addr >= 0xFE00 && addr <= 0xFE9F:
		return m.ppu.Read(addr)
	// [FF00-FF7F] Memory-mapped I/O
	case (addr >= 0xFF00 && addr <= 0xFF7F) || addr == 0xFFFF:
		if d := m.ioDevices[addr&0xFF]; d != nil {
			return d.Read(addr)
		}
		return 0xFF
	default:
		return 0xFF
	}
}

func (m *mmuImpl) Write(addr uint16, value byte) {
	// [FF80-FFFE] Zero-page RAM
	if addr >= 0xFF80 && addr < 0xFFFF {
		m.zpram[addr-0xFF80] = value
		return
	} else if m.dma.blockMemoryAccess(addr) {
		if addr == AddrDMATransfer {
			m.dma.Write(addr, value)
		}
		return
	}

	switch {
	// [0000-7FFF] Cartridge ROM
	case addr >= 0x0000 && addr <= 0x7FFF:
		m.cartridge.Write(addr, value)
	// [8000-9FFF] Graphics RAM
	case addr >= 0x8000 && addr <= 0x9FFF:
		m.ppu.Write(addr, value)
	// [A000-BFFF] Cartridge (External) RAM
	case addr >= 0xA000 && addr <= 0xBFFF:
		m.cartridge.Write(addr, value)
	// [C000-DFFF] Working RAM
	case addr >= 0xC000 && addr <= 0xDFFF:
		m.ram[addr-0xC000] = value
	// [E000-FDFF] Working RAM (shadow)
	case addr >= 0xE000 && addr <= 0xFDFF:
		m.ram[addr-0xE000] = value
	// [FE00-FE9F] Graphics: sprite information
	case addr >= 0xFE00 && addr <= 0xFE9F:
		m.ppu.Write(addr, value)
	// [FF00-FF7F] Memory-mapped I/O
	case (addr >= 0xFF00 && addr <= 0xFF7F) || addr == 0xFFFF:
		if d := m.ioDevices[addr&0xFF]; d != nil {
			d.Write(addr, value)
		}
	}
}

func (m *mmuImpl) RequestInterrupt(i IRQ) {
	m.Write(AddrIRQFlags, m.Read(AddrIRQFlags)|byte(i))
}

func (m *mmuImpl) GetCurrentIterrupt() IRQ {
	i := IRQ(m.Read(AddrIRQEnabled) & m.Read(AddrIRQFlags))
	handle := func(test IRQ) bool {
		if i&test == test {
			f := IRQ(m.Read(AddrIRQFlags))
			m.Write(AddrIRQFlags, byte(f&(0xFF^test)))
			return true
		}
		return false
	}

	switch {
	case handle(IRQVBlank):
		return IRQVBlank
	case handle(IRQLCDStat):
		return IRQLCDStat
	case handle(IRQTimer):
		return IRQTimer
	case handle(IRQSerial):
		return IRQSerial
	case handle(IRQJoypad):
		return IRQJoypad
	default:
		return IRQNone
	}

}
