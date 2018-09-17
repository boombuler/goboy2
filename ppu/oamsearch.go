package ppu

type oamSearch struct {
	oam oam

	readY     bool
	spriteX   byte
	spriteY   byte
	resIdx    int
	spriteIdx int
	sprites   []*spriteData
}

func (os *oamSearch) start(ppu *PPU) {
	os.readY = true
	os.spriteY = 0
	os.spriteX = 0
	os.resIdx = 0
	os.spriteIdx = 0
	if os.sprites == nil {
		os.sprites = make([]*spriteData, 10)
	} else {
		for i := 0; i < 10; i++ {
			os.sprites[i] = nil
		}
	}
}

func (os *oamSearch) state() ppuState {
	return sOAMRead
}

func (os *oamSearch) step(ppu *PPU) interface{} {
	if os.readY {
		os.spriteY = os.oam[os.spriteIdx].y
		os.readY = false
	} else {
		if os.resIdx < len(os.sprites) {
			os.spriteX = os.oam[os.spriteIdx].x
			var spriteHeight byte = 8
			if ppu.largeSprites() {
				spriteHeight = 16
			}

			if yTest := ppu.ly + 0x10; os.spriteY <= yTest && yTest < os.spriteY+spriteHeight {
				os.sprites[os.resIdx] = &os.oam[os.spriteIdx]
				os.resIdx++
			}
		}

		os.spriteIdx++
		os.readY = true
	}
	if os.spriteIdx >= 40 {
		return os.sprites
	}
	return nil
}
