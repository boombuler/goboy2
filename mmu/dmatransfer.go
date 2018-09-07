package mmu

type dmaTransfer struct {
	mmu   MMU
	steps uint16

	addr    uint16
	ticking bool
	lastVal byte
	block   bool
}

const (
	AddrDMATransfer uint16 = 0xFF46
	addrOAMStart    uint16 = 0xFE00
)

func (c *dmaTransfer) Read(addr uint16) byte {
	return c.lastVal
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
	i := 0x009F - c.steps
	c.ticking = true
	c.lastVal = c.mmu.Read(c.addr + i)
	c.mmu.Write(addrOAMStart+i, c.lastVal)
	c.ticking = false
}

func (c *dmaTransfer) blockMemoryAccess() bool {
	return !c.ticking && // get memory access while copy operation
		c.block
}
