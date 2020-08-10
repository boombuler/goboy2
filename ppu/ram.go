package ppu

type vRAM []byte

type spriteData struct {
	y      byte
	x      byte
	tileID byte
	flags  byte
}

type oam struct {
	data []spriteData
	gbc  bool
}

func newVRAM() vRAM {
	return make(vRAM, 0x2000)
}

func (vr vRAM) Read(addr uint16) byte {
	return vr[addr-0x8000]
}

func (vr vRAM) Write(addr uint16, val byte) {
	vr[addr-0x8000] = val
}

func newOAM(gbc bool) *oam {
	return &oam{make([]spriteData, 40), gbc}
}

func (sr oam) Read(addr uint16) byte {
	addr = (addr - 0xFE00) % 0x00A0
	no := addr >> 2
	switch addr & 0x03 {
	case 0:
		return sr.data[no].y
	case 1:
		return sr.data[no].x
	case 2:
		return sr.data[no].tileID
	case 3:
		return sr.data[no].flags
	}
	return 0x00
}

func (sr oam) Write(addr uint16, val byte) {
	addr = (addr - 0xFE00) % 0x00A0
	no := addr >> 2
	switch addr & 0x03 {
	case 0:
		sr.data[no].y = val
	case 1:
		sr.data[no].x = val
	case 2:
		sr.data[no].tileID = val
	case 3:
		sr.data[no].flags = val
	}
}

func (sp spriteData) palIdx(gbc bool) int {
	if gbc {
		return int(sp.flags & 0x03)
	}
	if int((sp.flags&0x10)>>4) == 0 {
		return 0
	}
	return 1
}

func (sp spriteData) vramHi() bool {
	return sp.flags&0x08 != 0
}

func (sp spriteData) flipH() bool {
	return (sp.flags & 0x20) != 0
}
func (sp spriteData) flipV() bool {
	return (sp.flags & 0x40) != 0
}
func (sp spriteData) prio() bool {
	return (sp.flags & 0x80) != 0
}
