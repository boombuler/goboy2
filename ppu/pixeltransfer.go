package ppu

type pixelTransfer struct {
	fifo    *pixelFiFo
	fetcher *fetcher

	dropped int
	curX    byte
	wnd     bool
}

func newPixelTransfer() *pixelTransfer {
	fifo := newPixelFiFo()
	return &pixelTransfer{
		fifo:    fifo,
		fetcher: newFetcher(fifo),
	}
}

func (pt *pixelTransfer) state() ppuState {
	return sPixelTransfer
}

func (pt *pixelTransfer) step(ppu *PPU) bool {
	pt.fetcher.tick(ppu)
	if ppu.useWndAndBg() {
		if pt.fifo.len <= 8 {
			return false
		}
		if pt.dropped < int(ppu.scrollX&8) {
			pt.fifo.dequeue(ppu) // drop pixel
			pt.dropped++
			return false
		}
		if !pt.wnd && ppu.useWnd() && ppu.ly >= ppu.winY && pt.curX == ppu.winX-7 {
			pt.wnd = true
			winX := uint16((int(pt.curX) - int(ppu.winX) + 7) / 8)
			winY := int(ppu.ly) - int(ppu.winY)
			mapAddr := ppu.wndTileMapDisplayAddr() + uint16((winY/0x08)*0x20)
			pt.fetcher.fetch(mapAddr, ppu.bgTileDataAddr(), winX, ppu.bgWndTileDataSigned(), byte(winY%0x08))
			return false
		}
	}

	if ppu.useObjects() {
		if pt.fetcher.isFetchingSprite() {
			return false
		}
		spriteAdded := false
		for i, s := range ppu.visibleSprites {
			if s == nil {
				continue
			}
			if pt.curX == 0 && s.x < 8 {
				if !spriteAdded {
					pt.fetcher.addSprite(ppu, s, 8-s.x)
					spriteAdded = true
				}
				ppu.visibleSprites[i] = nil
			} else if s.x-8 == pt.curX {
				if !spriteAdded {
					pt.fetcher.addSprite(ppu, s, 0)
					spriteAdded = true
				}
				ppu.visibleSprites[i] = nil
			}
			if spriteAdded {
				return false
			}
		}
	}

	color := pt.fifo.dequeue(ppu)
	ppu.curScreen.Set(int(pt.curX), int(ppu.ly), color)
	if pt.curX++; int(pt.curX) == DisplayWidth {
		return true
	}
	return false
}

func (pt *pixelTransfer) start(ppu *PPU) bool {
	if int(ppu.ly) >= DisplayHeight {
		return false
	}
	pt.wnd = false
	pt.dropped = 0
	pt.curX = 0

	pt.fetcher.reset()
	if ppu.useWndAndBg() {
		pt.fetcher.reset()
		bgX := uint16(ppu.scrollX) / 8
		bgY := (int(ppu.scrollY) + int(ppu.ly)) % 0x100
		mapAdr := ppu.bgTileDisplayAddr() + uint16((bgY/8)*0x20)
		pt.fetcher.fetch(mapAdr, ppu.bgTileDataAddr(), bgX, ppu.bgWndTileDataSigned(), byte(bgY%8))
	} else {
		pt.fetcher.disabled = true
	}
	return true
}
