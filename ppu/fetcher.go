package ppu

import (
	"fmt"
)

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
	tileAttr     tileAttr
	data1        byte
	data2        byte
	sprite       *spriteData
	spriteLine   byte
	spriteOffset byte
}

type tileAttr byte

func (ta tileAttr) palIdx(gbc bool) int {
	if gbc {
		return 0
	}
	return int(ta) & 0x07
}
func (ta tileAttr) vramHi() bool {
	return ta&(1<<3) != 0
}
func (ta tileAttr) flipH() bool {
	return ta&(1<<5) != 0
}
func (ta tileAttr) flipV() bool {
	return ta&(1<<6) != 0
}
func (ta tileAttr) prio() bool {
	return ta&(1<<7) != 0
}

type renderAttr interface {
	palIdx(gbc bool) int
	vramHi() bool
	flipH() bool
	flipV() bool
	prio() bool
}

var (
	emptyLine = []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
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
		addr := f.mapAddress + f.xOffset
		f.tileID = ppu.vram0.Read(addr)
		if ppu.dmgMode() {
			f.tileAttr = 0
		} else {
			f.tileAttr = tileAttr(ppu.vram1.Read(addr))
		}
		f.state = fsReadData1
	case fsReadData1:
		f.data1 = f.getTileData(ppu, f.tileLine, 0, f.tileAddress, f.signedIDs, 8, f.tileAttr)
		f.state = fsReadData2
		break
	case fsReadData2:
		f.data2 = f.getTileData(ppu, f.tileLine, 1, f.tileAddress, f.signedIDs, 8, f.tileAttr)
		f.state = fsPush
	case fsPush:
		if f.fifo.len <= 8 {
			f.fillPixBuffer(ppu, f.tileAttr, false)
			f.fifo.enqueue(f.pixBuffer)
			f.xOffset = (f.xOffset + 1) % 0x20
			f.state = fsReadTileID
		}
	case fsReadSpriteTileID:
		f.tileID = f.sprite.tileID
		h := ppu.spriteHeight()
		if h == 16 {
			f.tileID &= 0xFE
		}
		f.state = fsReadSpriteFlags
	case fsReadSpriteFlags:
		f.state = fsReadSpriteData1
	case fsReadSpriteData1:
		f.data1 = f.getTileData(ppu, f.spriteLine, 0, 0x8000, false, ppu.spriteHeight(), f.sprite)
		f.state = fsReadSpriteData2
		break

	case fsReadSpriteData2:
		f.data2 = f.getTileData(ppu, f.spriteLine, 1, 0x8000, false, ppu.spriteHeight(), f.sprite)
		f.state = fsPushSprite
		break

	case fsPushSprite:
		f.fillPixBuffer(ppu, f.sprite, true)
		f.fifo.setOverlay(f.pixBuffer, int(f.spriteOffset))
		f.state = fsReadTileID
		break
	}
}

func (f *fetcher) getTileData(ppu *PPU, line byte, byteNumber byte, tileDataAddress uint16, signed bool, tileHeight byte, attrs renderAttr) byte {
	if attrs.flipV() {
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
	if attrs.vramHi() {
		return ppu.vram1.Read(addr)
	}
	return ppu.vram0.Read(addr)
}

func (f *fetcher) fillPixBuffer(ppu *PPU, attr renderAttr, isObj bool) {
	ifBitThen := func(src byte, bit byte, hiVal byte) byte {
		if ((src >> bit) & 1) != 0 {
			return hiVal
		}
		return 0
	}

	palIdx := attr.palIdx(!ppu.dmgMode()) & 0x07
	if palIdx != 0 && ppu.dmgMode() {
		fmt.Println("blub")
	}
	pixBase := byte(palIdx) << 4
	if !isObj {
		pixBase |= 0x80
	}
	if attr.prio() && (isObj || ppu.masterPriority()) {
		pixBase |= 0x04
	}

	for i := 7; i >= 0; i-- {
		p := pixBase | ifBitThen(f.data1, byte(i), 1) | ifBitThen(f.data2, byte(i), 2)

		if attr.flipH() {
			f.pixBuffer[i] = p
		} else {
			f.pixBuffer[7-i] = p
		}
	}
}
