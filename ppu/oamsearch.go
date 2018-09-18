package ppu

type oamSearch struct {
	readY     bool
	spriteX   byte
	spriteY   byte
	resIdx    int
	spriteIdx int
}

func (os *oamSearch) start(ppu *PPU) bool {
	if int(ppu.ly) >= DisplayHeight {
		return false
	}
	os.readY = true
	os.spriteY = 0
	os.spriteX = 0
	os.resIdx = 0
	os.spriteIdx = 0
	if ppu.visibleSprites == nil {
		ppu.visibleSprites = make([]*spriteData, 10)
	} else {
		for i := 0; i < 10; i++ {
			ppu.visibleSprites[i] = nil
		}
	}
	ppu.requstLcdcInterrupt(liOAM)
	return true
}

func (os *oamSearch) state() ppuState {
	return sOAMRead
}

func (os *oamSearch) step(ppu *PPU) bool {
	if os.readY {
		os.spriteY = ppu.oam[os.spriteIdx].y
		os.readY = false
	} else {
		if os.resIdx < len(ppu.visibleSprites) {
			os.spriteX = ppu.oam[os.spriteIdx].x
			var spriteHeight byte = 8
			if ppu.largeSprites() {
				spriteHeight = 16
			}

			if yTest := ppu.ly + 0x10; os.spriteY <= yTest && yTest < os.spriteY+spriteHeight {
				ppu.visibleSprites[os.resIdx] = &ppu.oam[os.spriteIdx]
				os.resIdx++
			}
		}

		os.spriteIdx++
		os.readY = true
	}
	if os.spriteIdx >= 40 {
		return true
	}
	return false
}