package mmu

type IRQ byte

const (
	IRQNone    IRQ = 0x00
	IRQVBlank  IRQ = 0x01
	IRQLCDStat IRQ = 0x02
	IRQTimer   IRQ = 0x04
	IRQSerial  IRQ = 0x08
	IRQJoypad  IRQ = 0x10
	IRQAll         = IRQVBlank | IRQLCDStat | IRQTimer | IRQSerial | IRQJoypad

	AddrIRQEnabled uint16 = 0xFFFF
	AddrIRQFlags   uint16 = 0xFF0F
)

type irqHandler struct {
	flag, mask IRQ
}

func (h *irqHandler) Read(addr uint16) byte {
	switch addr {
	case AddrIRQEnabled:
		return byte(h.mask)
	case AddrIRQFlags:
		return byte(h.flag)
	default:
		return 0
	}
}
func (h *irqHandler) Write(addr uint16, value byte) {
	switch addr {
	case AddrIRQEnabled:
		h.mask = IRQ(value)
	case AddrIRQFlags:
		h.flag = IRQ(value)
	default:
		return
	}
}

func (i IRQ) Address() uint16 {
	switch i {
	case IRQNone:
		return 0x0000
	case IRQVBlank:
		return 0x0040
	case IRQLCDStat:
		return 0x0048
	case IRQTimer:
		return 0x0050
	case IRQSerial:
		return 0x0058
	case IRQJoypad:
		return 0x0060
	}
	panic("Invalid IRQ")
}
