package mmu

type dmaTransfer struct {
	mmu   MMU
	steps uint16

	addr    uint16
	ticking bool
	block   bool
}

const (
	addrOAMStart uint16 = 0xFE00
)

func (c *dmaTransfer) Read(addr uint16) byte {
	return byte(c.addr >> 8)
}

func (c *dmaTransfer) Write(addr uint16, value byte) {
	c.addr = uint16(value) << 8 & 0xFF00
	c.steps = 162
}

func (c *dmaTransfer) step() {
	if c.steps == 0 {
		return
	}
	c.steps--
	if c.steps == 160 {
		c.block = true
	} else if c.steps == 0 {
		c.block = false
	}

	if c.steps >= 160 {
		return // warmup
	}
	i := 159 - c.steps
	srcAdr := c.addr + i
	destAdr := addrOAMStart + i

	c.ticking = true
	c.mmu.Write(destAdr, c.mmu.Read(srcAdr))
	c.ticking = false
}

type memBus byte

const (
	busCPU memBus = iota
	busMain
	busVRam
)

var memBusMap = [8]memBus{
	busMain, // 0x0000
	busMain, // 0x2000
	busMain, // 0x4000
	busMain, // 0x6000
	busVRam, // 0x8000
	busMain, // 0xA000
	busMain, // 0xC000
	busCPU,  // 0xE000
}

func busFromAdr(addr uint16) memBus {
	return memBusMap[int(addr>>13)]
}

func (c *dmaTransfer) blockMemoryAccess(addr uint16) bool {
	if !c.ticking && c.block {
		// Always block oam
		if addr >= 0xFE00 && addr <= 0xFE9F {
			return true
		}
		return busFromAdr(c.addr) == busFromAdr(addr)
	}
	return false
}
