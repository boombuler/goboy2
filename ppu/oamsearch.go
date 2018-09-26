package ppu

import "goboy2/consts"

type oamSearch struct {
	readY     bool
	spriteX   byte
	spriteY   byte
	resIdx    int
	spriteIdx int
}

const visibleSpriteDataCount = 10

func (os *oamSearch) start(ppu *PPU) bool {
	if int(ppu.ly) >= consts.DisplayHeight {
		return false
	}
	os.readY = true
	os.spriteY = 0
	os.spriteX = 0
	os.resIdx = 0
	os.spriteIdx = 0
	if ppu.visibleSprites == nil {
		ppu.visibleSprites = make([]*spriteData, visibleSpriteDataCount)
	} else {
		for i := 0; i < visibleSpriteDataCount; i++ {
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
		os.spriteY = ppu.oam.data[os.spriteIdx].y
		os.readY = false
	} else {
		if os.resIdx < len(ppu.visibleSprites) {
			os.spriteX = ppu.oam.data[os.spriteIdx].x

			if yTest := ppu.ly + 0x10; os.spriteY <= yTest && yTest < os.spriteY+ppu.spriteHeight() {
				ppu.visibleSprites[os.resIdx] = &ppu.oam.data[os.spriteIdx]
				os.resIdx++
			}
		}

		os.spriteIdx++
		os.readY = true
	}
	if os.spriteIdx >= len(ppu.oam.data) {
		return true
	}
	return false
}
