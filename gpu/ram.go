package gpu

type spriteData struct {
	Y        int
	X        int
	TileId   int
	Priority bool
	FlipV    bool
	FlipH    bool
	Palette  int
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
		val := byte(0)
		if sr[no].Priority {
			val |= 0x80
		}
		if sr[no].FlipV {
			val |= 0x40
		}
		if sr[no].FlipH {
			val |= 0x20
		}
		val |= byte(sr[no].Palette << 4)
		return val
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
		sr[no].Priority = (val & 0x80) != 0x80
		sr[no].FlipV = (val & 0x40) == 0x40
		sr[no].FlipH = (val & 0x20) == 0x20
		sr[no].Palette = int((val & 0x10) >> 4)
	}
}
