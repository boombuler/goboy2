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
	skipTick  bool
	fifo      *pixelFiFo
	pixBuffer []byte

	mapAddress   uint16
	tileAddress  uint16
	xOffset      uint16
	signedIDs    bool
	tileLine     byte
	tileID       byte
	tileAttr     byte
	data1        byte
	data2        byte
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
	f.skipTick = true
}

func (f *fetcher) fetchSprite(ppu *PPU, s *spriteData, offset byte) {
	f.sprite = s
	f.spriteLine = ppu.ly + 16 - s.y
	f.spriteOffset = offset
	f.state = fsReadSpriteTileID
	f.skipTick = true
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
		f.tileID = ppu.vram0.Read(f.mapAddress + f.xOffset)
		if ppu.mmu.GBC() {
			f.tileAttr = ppu.vram1.Read(f.mapAddress + f.xOffset)
		} else {
			f.tileAttr = 0
		}
		f.state = fsReadData1
	case fsReadData1:
		flipV := f.tileAttr&(1<<6) != 0
		hiBank := f.tileAttr&(1<<3) != 0
		f.data1 = f.getTileData(ppu, hiBank, f.tileLine, 0, f.tileAddress, f.signedIDs, flipV, 8)
		f.state = fsReadData2
		break
	case fsReadData2:
		flipV := f.tileAttr&(1<<6) != 0
		hiBank := f.tileAttr&(1<<3) != 0
		f.data2 = f.getTileData(ppu, hiBank, f.tileLine, 1, f.tileAddress, f.signedIDs, flipV, 8)
		f.state = fsPush
	case fsPush:
		if f.fifo.len <= 8 {
			flipH := f.tileAttr&(1<<5) != 0
			prio := f.tileAttr&(1<<7) != 0
			f.fillPixBuffer(flipH, prio, psBG, int(f.tileAttr&0x03))
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
		h := ppu.spriteHeight()
		if h == 16 {
			f.tileID &= 0xFE
		}
		f.data1 = f.getTileData(ppu, f.sprite.vramHi(), f.spriteLine, 0, 0x8000, false, f.sprite.flipV(), h)
		f.state = fsReadSpriteData2
		break

	case fsReadSpriteData2:
		f.data2 = f.getTileData(ppu, f.sprite.vramHi(), f.spriteLine, 1, 0x8000, false, f.sprite.flipV(), ppu.spriteHeight())
		f.state = fsPushSprite
		break

	case fsPushSprite:
		f.fillPixBuffer(f.sprite.flipH(), f.sprite.priority(), psObj, f.sprite.palette(ppu.mmu.GBC()))
		f.fifo.setOverlay(f.pixBuffer, int(f.spriteOffset))
		f.state = fsReadTileID
		break
	}
}

func (f *fetcher) getTileData(ppu *PPU, hiBank bool, line byte, byteNumber byte, tileDataAddress uint16, signed bool, flipV bool, tileHeight byte) byte {
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
	addr := tileAddress + uint16(line*2+byteNumber)
	if hiBank {
		return ppu.vram1.Read(addr)
	}
	return ppu.vram0.Read(addr)
}

func (f *fetcher) fillPixBuffer(flipX bool, priority bool, src paletteSrc, pIdx int) {
	ifBitThen := func(src byte, bit byte, hiVal byte) byte {
		if ((src >> bit) & 1) != 0 {
			return hiVal
		}
		return 0
	}

	for i := 7; i >= 0; i-- {
		p := ifBitThen(f.data2, byte(i), 2) | ifBitThen(f.data1, byte(i), 1)
		if priority {
			p |= 0x04
		}
		if src == psBG {
			p |= 0x80
		}
		p |= byte(pIdx&7) << 4

		if flipX {
			f.pixBuffer[i] = p
		} else {
			f.pixBuffer[7-i] = p
		}
	}
}
