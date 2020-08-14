package mmu

import (
	"github.com/boombuler/goboy2/cartridge"
	"github.com/boombuler/goboy2/consts"
)

type MMU interface {
	IODevice
	HardwareCompat() consts.HardwareCompat
	EmuMode() consts.HardwareCompat
	RequestInterrupt(i IRQ)
	GetCurrentIterrupt() IRQ
	ConnectPPU(ppu IODevice)
	LoadCartridge(cartridge *cartridge.Cartridge)
	AddIODevice(d IODevice, addrs ...uint16)
	Step()
	Init(noBoot bool)
}

type mmuImpl struct {
	hw        consts.HardwareCompat
	ioDevices []IODevice
	cartridge *cartridge.Cartridge
	ppu       IODevice
	ram       IODevice
	zpram     [127]byte
	dma       *dmaTransfer
	lcdMode   byte
}

type IODevice interface {
	Read(addr uint16) byte
	Write(addr uint16, value byte)
}

type IOAddrDevice interface {
	IOAddrs() []uint16
}

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

func New(hw consts.HardwareCompat) MMU {
	res := &mmuImpl{
		hw:        hw,
		ioDevices: make([]IODevice, 256),
	}
	res.ram = newWorkingRAM(res)
	res.dma = &dmaTransfer{mmu: res}
	res.AddIODevice(new(irqHandler), consts.AddrIRQFlags, consts.AddrIRQEnabled)
	res.AddIODevice(new(bootMode), consts.AddrBootmodeFlag)
	res.AddIODevice(res.dma, consts.AddrDMATransfer)
	if hw == consts.GBC {
		// Add undocumented GBC registers.
		res.AddIODevice(newGBCRegisters(res))
	}

	return res
}

func (m *mmuImpl) Step() {
	m.dma.step()
}

func (m *mmuImpl) AddIODevice(d IODevice, addrs ...uint16) {
	if addrIO, ok := d.(IOAddrDevice); ok && len(addrs) == 0 {
		addrs = addrIO.IOAddrs()
	}

	for _, address := range addrs {
		m.ioDevices[address&0xFF] = d
	}
}

func (m *mmuImpl) HardwareCompat() consts.HardwareCompat {
	return m.hw
}

func (m *mmuImpl) EmuMode() consts.HardwareCompat {
	if m.hw == consts.GBC && (m.lcdMode != 4 || m.Read(consts.AddrBootmodeFlag) == 0x00) {
		return consts.GBC
	}
	return consts.DMG
}

func (m *mmuImpl) LoadCartridge(cartridge *cartridge.Cartridge) {
	m.cartridge = cartridge
}

func (m *mmuImpl) ConnectPPU(ppu IODevice) {
	m.ppu = ppu
}

func (m *mmuImpl) Init(noBoot bool) {
	if noBoot {
		if !m.cartridge.GBC {
			m.lcdMode = 4
		}
		m.Write(consts.AddrBootmodeFlag, 0x01) // Disable Boot ROM.
		m.Write(consts.AddrIRQFlags, 1)
	}
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
		if m.Read(consts.AddrBootmodeFlag) == 0x00 {
			switch m.hw {
			case consts.GBC:
				// if in gbc mode and reading Cartridge header then read from card...
				if addr < uint16(len(GBC_BOOTROM)) && (addr < 0x0100 || addr > 0x014F) {
					return GBC_BOOTROM[addr]
				}
			case consts.DMG:
				if addr < uint16(len(BOOTROM)) {
					return BOOTROM[addr]
				}
			}
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
	// [C000-FDFF] Working RAM
	case addr >= 0xC000 && addr <= 0xFDFF:
		return m.ram.Read(addr)
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
		if addr == consts.AddrDMATransfer {
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
	// [C000-FDFF] Working RAM
	case addr >= 0xC000 && addr <= 0xFDFF:
		m.ram.Write(addr, value)
	// [FE00-FE9F] Graphics: sprite information
	case addr >= 0xFE00 && addr <= 0xFE9F:
		m.ppu.Write(addr, value)
	// [FF00-FF7F] Memory-mapped I/O
	case (addr >= 0xFF00 && addr <= 0xFF7F) || addr == 0xFFFF:
		if addr == consts.AddrLCDMODE && m.HardwareCompat() == consts.GBC && m.Read(consts.AddrBootmodeFlag) == 0x00 {
			m.lcdMode = value
		}

		if d := m.ioDevices[addr&0xFF]; d != nil {
			d.Write(addr, value)
		}
	}
}

func (m *mmuImpl) RequestInterrupt(i IRQ) {
	m.Write(consts.AddrIRQFlags, m.Read(consts.AddrIRQFlags)|byte(i))
}

func (m *mmuImpl) GetCurrentIterrupt() IRQ {
	i := IRQ(m.Read(consts.AddrIRQEnabled) & m.Read(consts.AddrIRQFlags))
	handle := func(test IRQ) bool {
		if i&test == test {
			f := IRQ(m.Read(consts.AddrIRQFlags))
			m.Write(consts.AddrIRQFlags, byte(f&(0xFF^test)))
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
