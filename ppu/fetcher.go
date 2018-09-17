package ppu

type fetcherState byte

const (
	fsReadTileID fetcherState = iota
	fsReadData1
	fsReadData2
	fsPush
	fsReadSpriteTileID
	fsReadSpriteFlags
	fsReadSpriteData1
	fsReadSpriteData2
	fsPushSprite
)

type fetcher struct {
	disabled bool
	state    fetcherState
	fifo     *pixelFiFo

	mapAddress  uint16
	tileAddress uint16
	xOffset     uint16
	signedAddrs bool
	tileLine    int
	tileID      byte
	tileData1   byte
	tileData2   byte
	skipTick    bool
}

func newFetcher(fifo *pixelFiFo) *fetcher {
	return &fetcher{
		fifo: fifo,
	}
}

func (f *fetcher) reset() {

}

func (f *fetcher) fetch(mapAddress uint16, tileAddress uint16, xOffset uint16, signedAddrs bool, tileLine int) {
	f.mapAddress = mapAddress
	f.tileAddress = tileAddress
	f.xOffset = xOffset
	f.signedAddrs = signedAddrs
	f.tileLine = tileLine
	f.fifo.clear()

	f.state = fsReadTileID
	f.tileID = 0
	f.tileData1 = 0
	f.tileData2 = 0
	f.skipTick = true // will be toggled on tick
}

func (f *fetcher) tick(ppu *PPU) {
	if f.disabled && f.state == fsReadTileID {
		if f.fifo.len() <= 8 {
			f.fifo.enqueue8Pixels(EMPTY_PIXEL_LINE)
		}
		return
	}

	f.skipTick = !f.skipTick
	if f.skipTick {
		return
	}

	switch f.state {
	case fsReadTileID:
		f.tileID = ppu.vram.Read(f.mapAddress + f.xOffset)
		f.state = fsReadData1
	case fsReadData1:
		f.tileData1 = getTileData(tileId, tileLine, 0, tileDataAddress, tileIdSigned, tileAttributes, 8)
		f.state = fsReadData2
		break

	case fsReadData2:
		f.tileData2 = getTileData(tileId, tileLine, 1, tileDataAddress, tileIdSigned, tileAttributes, 8)
		f.state = fsPush

	case fsPush:
		if fifo.getLength() <= 8 {
			f.fifo.enqueue8Pixels(zip(tileData1, tileData2, tileAttributes.isXflip()), tileAttributes)
			f.xOffset = (xOffset + 1) % 0x20
			f.state = fsReadTileID
		}
		break

	case fsReadSpriteTileId:
		f.tileId = sprite.tileID
		f.state = fsReadSpriteFlags
		break

	case fsReadSpriteFlags:
		f.spriteAttributes = TileAttributes.valueOf(oemRam.getByte(sprite.getAddress() + 3))
		f.state = fsReadSpriteData1
		break

	case fsReadSpriteData1:
		if lcdc.getSpriteHeight() == 16 {
			tileId &= 0xfe
		}
		f.tileData1 = getTileData(tileId, spriteTileLine, 0, 0x8000, false, spriteAttributes, lcdc.getSpriteHeight())
		f.state = fsReadSpriteData2
		break

	case fsReadSpriteData2:
		f.tileData2 = getTileData(tileId, spriteTileLine, 1, 0x8000, false, spriteAttributes, lcdc.getSpriteHeight())
		f.state = fsPushSprite
		break

	case fsPushSprite:
		f.fifo.setOverlay(zip(tileData1, tileData2, spriteAttributes.isXflip()), spriteOffset, spriteAttributes, spriteOamIndex)
		f.state = fsReadTileID
		break
	}
}

func (f *fetcher) getTileData(int tileId, int line, int byteNumber, int tileDataAddress, boolean signed, TileAttributes attr, int tileHeight) int {
	int effectiveLine;
	if (attr.isYflip()) {
		effectiveLine = tileHeight - 1 - line;
	} else {
		effectiveLine = line;
	}

	int tileAddress;
	if (signed) {
		tileAddress = tileDataAddress + toSigned(tileId) * 0x10;
	} else {
		tileAddress = tileDataAddress + tileId * 0x10;
	}
	AddressSpace videoRam = (attr.getBank() == 0 || !gbc) ? videoRam0 : videoRam1;
	return videoRam.getByte(tileAddress + effectiveLine * 2 + byteNumber);
}