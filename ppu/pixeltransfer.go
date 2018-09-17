package ppu

type pixelTransfer struct {
	fifo    *pixelFiFo
	fetcher *fetcher
	sprites []*spriteData

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

func (pt *pixelTransfer) step(ppu *PPU) interface{} {

}

func (pt *pixelTransfer) start(ppu *PPU) {
	pt.sprites = ppu.lastPhaseResult.([]*spriteData)
	pt.wnd = false
	pt.dropped = 0
	pt.curX = 0

	pt.fetcher.reset()
	if ppu.useWndAndBg() {
		bgX := int(ppu.scrollX) / 8
		bgY := (int(ppu.scrollY) + int(ppu.ly)) % 0x100
		mapAdr := ppu.bgTileDisplayAddr() + uint16((bgY/8)*0x20)
		pt.fetcher.fetch(mapAdr, ppu.bgTileDataAddr(), bgX, ppu.bgWndTileDataSigned(), bgY%8)
	} else {
		pt.fetcher.disabled = true
	}
}

func (pt *pixelTransfer) startFetchingWindow(ppu *PPU) {
	winX := (int(pt.curX) - int(ppu.winX) + 7) / 8
	winY := int(ppu.ly) - int(ppu.winY)

	mapAddr := ppu.wndTileMapDisplayAddr() + uint16((winY/0x08)*0x20)

	pt.fetcher.fetch(mapAddr, ppu.bgTileDataAddr(), winX, ppu.bgWndTileDataSigned(), winY%0x08)
}
