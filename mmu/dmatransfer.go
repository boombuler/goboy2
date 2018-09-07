package mmu

type dmaTransfer struct {
	mmu MMU
}

const AddrDMATransfer uint16 = 0xFF46

func (c *dmaTransfer) Read(addr uint16) byte {
	return 0x00
}

func (c *dmaTransfer) Write(addr uint16, value byte) {
	from := uint16(value) << 8 & 0xFF00
	to := uint16(0xFE00)
	for i := uint16(0); i <= 0x009F; i++ {
		c.mmu.Write(to+i, c.mmu.Read(from+i))
	}
}
