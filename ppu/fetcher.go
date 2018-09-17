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
	disabled  bool
	state     fetcherState
	fifo      *pixelFiFo
	pixBuffer []byte

	mapAddress   uint16
	tileAddress  uint16
	xOffset      uint16
	signedIDs    bool
	tileLine     byte
	tileID       byte
	tileData1    byte
	tileData2    byte
	skipTick     bool
	sprite       *spriteData
	spriteLine   byte
	spriteOffset byte
}

var (
	emptyLine = []byte{0, 0, 0, 0, 0, 0, 0, 0}
)

func newFetcher(fifo *pixelFiFo) *fetcher {
	return &fetcher{
		fifo:      fifo,
		pixBuffer: make([]byte, 8),
	}
}

func (f *fetcher) reset() {
	f.disabled = false
	f.state = fsReadTileID
	f.tileID = 0
	f.tileData1 = 0
	f.tileData2 = 0
	f.skipTick = true // will be toggled on tick
}

func (f *fetcher) addSprite(ppu *PPU, s *spriteData, offset byte) {
	f.sprite = s
	f.spriteLine = ppu.ly + 16 - s.y
	f.spriteOffset = offset
	f.state = fsReadSpriteTileID
	f.skipTick = true // will be toggled on tick
}

func (f *fetcher) isFetchingSprite() bool {
	return f.state >= fsReadSpriteTileID && f.state <= fsPushSprite
}

func (f *fetcher) fetch(mapAddress uint16, tileAddress uint16, xOffset uint16, signedIDs bool, tileLine byte) {
	f.mapAddress = mapAddress
	f.tileAddress = tileAddress
	f.xOffset = xOffset
	f.signedIDs = signedIDs
	f.tileLine = tileLine
	f.fifo.clear()

	f.state = fsReadTileID
	f.skipTick = true // will be toggled on tick
}

func (f *fetcher) tick(ppu *PPU) {
	if f.disabled && f.state == fsReadTileID {
		if f.fifo.len <= 8 {
			f.fifo.enqueue(emptyLine)
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
		f.tileData1 = f.getTileData(ppu, f.tileLine, 0, f.tileAddress, f.signedIDs, false, 8)
		f.state = fsReadData2
		break

	case fsReadData2:
		f.tileData2 = f.getTileData(ppu, f.tileLine, 1, f.tileAddress, f.signedIDs, false, 8)
		f.state = fsPush

	case fsPush:
		if f.fifo.len <= 8 {
			f.fillPixBuffer(false, false, psBG)
			f.fifo.enqueue(f.pixBuffer)
			f.xOffset = (f.xOffset + 1) % 0x20
			f.state = fsReadTileID
		}
	case fsReadSpriteTileID:
		f.tileID = f.sprite.tileID
		f.state = fsReadSpriteFlags
	case fsReadSpriteFlags:
		f.state = fsReadSpriteData1
	case fsReadSpriteData1:
		h := byte(8)
		if ppu.largeSprites() {
			f.tileID &= 0xFE
			h = 16
		}
		f.tileData1 = f.getTileData(ppu, f.spriteLine, 0, 0x8000, false, f.sprite.flipV(), h)
		f.state = fsReadSpriteData2
		break

	case fsReadSpriteData2:
		h := byte(8)
		if ppu.largeSprites() {
			f.tileID &= 0xFE
			h = 16
		}
		f.tileData2 = f.getTileData(ppu, f.spriteLine, 1, 0x8000, false, f.sprite.flipV(), h)
		f.state = fsPushSprite
		break

	case fsPushSprite:
		p := psObj0
		if f.sprite.palette() == 1 {
			p = psObj1
		}

		f.fillPixBuffer(f.sprite.flipH(), f.sprite.priority(), p)
		f.fifo.setOverlay(f.pixBuffer, int(f.spriteOffset))
		f.state = fsReadTileID
		break
	}
}

func (f *fetcher) getTileData(ppu *PPU, line byte, byteNumber byte, tileDataAddress uint16, signed bool, flipV bool, tileHeight byte) byte {
	if flipV {
		line = tileHeight - 1 - line
	}

	var tileAddress uint16
	if signed {
		signedID := int8(f.tileID)
		tileAddress = uint16(int(tileDataAddress) + int(signedID)*0x10)
	} else {
		tileAddress = tileDataAddress + uint16(f.tileID)*0x10
	}
	return ppu.vram.Read(tileAddress + uint16(line*2+byteNumber))
}

func (f *fetcher) fillPixBuffer(flipX bool, priority bool, src paletteSrc) {
	ifBitThen := func(src byte, bit byte, hiVal byte, loVal byte) byte {
		if ((src >> bit) & 1) != 0 {
			return hiVal
		}
		return loVal
	}

	for i := 7; i >= 0; i-- {
		p := ifBitThen(f.tileData2, byte(i), 2, 0) | ifBitThen(f.tileData1, byte(i), 1, 0)
		if priority {
			p |= 0x80
		}
		p |= (byte(src) << 4)

		if flipX {
			f.pixBuffer[i] = p
		} else {
			f.pixBuffer[7-i] = p
		}
	}
}
