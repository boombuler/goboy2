package gpu

type spriteData struct {
	Y      int
	X      int
	TileId int
	flags  byte
}
type vRAM []byte
type oam []*spriteData

func newVRam() vRAM {
	return make(vRAM, 0x2000)
}

func newOAM() oam {
	res := make(oam, 0x00A0)
	for i := range res {
		res[i] = new(spriteData)
	}
	return res
}

func (vr vRAM) Read(addr uint16) byte {
	return vr[addr-0x8000]
}

func (vr vRAM) Write(addr uint16, val byte) {
	vr[addr-0x8000] = val
}

func (sr oam) Read(addr uint16) byte {
	addr = (addr - 0xFE00) % 0x00A0
	no := addr >> 2
	switch addr & 0x03 {
	case 0:
		return byte(sr[no].Y + 16)
	case 1:
		return byte(sr[no].X + 8)
	case 2:
		return byte(sr[no].TileId)
	case 3:
		return sr[no].flags
	}
	return 0x00
}

func (sr oam) Write(addr uint16, val byte) {
	addr = (addr - 0xFE00) % 0x00A0
	no := addr >> 2
	switch addr & 0x03 {
	case 0:
		sr[no].Y = int(val) - 16
	case 1:
		sr[no].X = int(val) - 8
	case 2:
		sr[no].TileId = int(val)
	case 3:
		sr[no].flags = val
	}
}

func (sp *spriteData) Palette() int {
	return int((sp.flags & 0x10) >> 4)
}
func (sp *spriteData) FlipH() bool {
	return (sp.flags & 0x20) != 0
}
func (sp *spriteData) FlipV() bool {
	return (sp.flags & 0x40) != 0
}
func (sp *spriteData) Priority() bool {
	return (sp.flags & 0x80) != 0
}
