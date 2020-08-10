package ppu

import "github.com/boombuler/goboy2/consts"

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

	if pt.fifo.len <= 8 {
		return false
	}

	if ppu.useWndAndBg() {
		if pt.dropped < int(ppu.scrollX%8) {
			pt.fifo.dequeue(ppu)
			pt.dropped++
			return false
		}
		if !pt.wnd && ppu.useWnd() && ppu.ly >= ppu.winY && (pt.curX + 7) >= ppu.winX {
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
		for i, s := range ppu.visibleSprites {
			if s < 0 {
				continue
			}
			sprite := ppu.oam.data[s]
			if pt.curX == 0 && sprite.x < 8 {
				pt.fetcher.fetchSprite(ppu, s, 8-sprite.x)
				ppu.visibleSprites[i] = -1
				return false

			} else if sprite.x-8 == pt.curX {
				pt.fetcher.fetchSprite(ppu, s, 0)
				ppu.visibleSprites[i] = -1
				return false
			}
		}

	}

	color := pt.fifo.dequeue(ppu)
	ppu.curScreen[int(ppu.ly)*consts.DisplayWidth+int(pt.curX)] = color
	if pt.curX++; int(pt.curX) == consts.DisplayWidth {
		return true
	}
	return false
}

func (pt *pixelTransfer) start(ppu *PPU) bool {
	if int(ppu.ly) >= consts.DisplayHeight {
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
