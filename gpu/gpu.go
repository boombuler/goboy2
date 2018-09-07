package gpu

import (
	"goboy2/mmu"
	"image"
	"image/color"
)

const (
	LCDC            uint16 = 0xFF40
	STAT                   = 0xFF41
	SCROLLY                = 0xFF42
	SCROLLX                = 0xFF43
	LY                     = 0xFF44
	LYC                    = 0xFF45
	BGP                    = 0xFF47
	OBJECTPALETTE_0        = 0xFF48
	OBJECTPALETTE_1        = 0xFF49
	WY                     = 0xFF4A
	WX                     = 0xFF4B

	HBLANK   byte = 0x00
	VBLANK   byte = 0x01
	OAMREAD  byte = 0x02
	VRAMREAD byte = 0x03

	TILEMAP0 uint16 = 0x9800
	TILEMAP1        = 0x9C00
)

type RawTile [16]byte
type Tile [8][8]int
type Palette [4]color.Color

type GPU struct {
	mmu.MMU
	screen *image.RGBA

	vram vRAM
	oam  oam

	bgTilemap     uint16
	windowTilemap uint16

	largeSprites bool
	spritesOn    bool
	bgOn         bool
	windowOn     bool
	tiledata0    bool

	bgPalette      Palette
	objectPalettes [2]Palette

	vBlankInterruptThrown bool
	lcdInterruptThrown    bool
	displayOn             bool
	clock                 int
	mode                  byte
	lcdc                  byte
	stat                  byte
	scrollx               byte
	scrolly               byte
	ly                    int
	lyc                   byte
	bgp                   byte
	op0                   byte
	op1                   byte
	wy                    byte
	wx                    byte
}

func (g *GPU) Read(addr uint16) byte {
	switch addr {
	case LCDC:
		return g.lcdc
	case STAT:
		return g.mode | g.stat&0xF8
	case SCROLLY:
		return g.scrolly
	case SCROLLX:
		return g.scrollx
	case LY:
		return byte(g.ly)
	case LYC:
		return g.lyc
	case BGP:
		return g.bgp
	case OBJECTPALETTE_0:
		return g.op0
	case OBJECTPALETTE_1:
		return g.op1
	case WX:
		return g.wx
	case WY:
		return g.wy
	default:
		return 0x00
	}
}

func (g *GPU) Write(addr uint16, value byte) {
	switch addr {
	case LCDC:
		g.lcdc = value

		g.displayOn = value&0x80 == 0x80

		if value&0x40 == 0x40 { //bit 6
			g.windowTilemap = TILEMAP1
		} else {
			g.windowTilemap = TILEMAP0
		}

		g.windowOn = value&0x20 == 0x20  //bit 5
		g.tiledata0 = value&0x10 != 0x10 //bit 4

		if value&0x08 == 0x08 { //bit 3
			g.bgTilemap = TILEMAP1
		} else {
			g.bgTilemap = TILEMAP0
		}

		g.largeSprites = value&0x04 == 0x04
		g.spritesOn = value&0x02 == 0x02 //bit 1
		g.bgOn = value&0x01 == 0x01      //bit 0
	case STAT:
		g.stat = (g.stat & 0x0F) | value
	case SCROLLY:
		g.scrolly = value
	case SCROLLX:
		g.scrollx = value
	case WX:
		g.wx = value
	case WY:
		g.wy = value
	case LY:
		g.ly = 0
	case LYC:
		g.lyc = value
	case BGP:
		g.bgp = value
		g.bgPalette = g.byteToPalette(value)
	case OBJECTPALETTE_0:
		g.op0 = value
		g.objectPalettes[0] = g.byteToPalette(value)
	case OBJECTPALETTE_1:
		g.op1 = value
		g.objectPalettes[1] = g.byteToPalette(value)
	}
}

func New(m mmu.MMU) *GPU {
	gpu := new(GPU)
	gpu.MMU = m
	gpu.clock = 456
	gpu.screen = newScreen()
	gpu.vram, gpu.oam = newVRam(), newOAM()
	m.SetGraphicRam(gpu.vram, gpu.oam)
	m.AddIODevice(gpu, LCDC, STAT, SCROLLY, SCROLLX, LY, LYC, BGP, OBJECTPALETTE_0, OBJECTPALETTE_1, WY, WX)
	gpu.Write(LCDC, 0x80)
	return gpu
}

func (g *GPU) Step(t uint16) (result *image.RGBA) {
	result = nil
	if !g.displayOn {
		g.ly = 0
		g.clock = 456
		g.mode = HBLANK
		return
	} else {
		if g.ly >= 144 {
			g.mode = VBLANK
			g.lcdInterruptThrown = false
		} else if g.clock >= 456-80 {
			g.mode = OAMREAD
			g.lcdInterruptThrown = false
		} else if g.clock >= 456-80-172 {
			g.mode = VRAMREAD
			g.lcdInterruptThrown = false
		} else {
			g.mode = HBLANK
			//throw HBlank LCD interrupt (if enabled)
			if g.HblankLCDInterruptEnabled() && !g.lcdInterruptThrown {
				g.RequestInterrupt(mmu.IRQLCDStat)
				g.lcdInterruptThrown = true
			}
		}
	}

	g.clock -= int(t)

	if g.clock <= 0 {
		g.clock += 456
		g.ly += 1

		if g.ly == 144 {
			//throw vblank interrupt
			if !g.vBlankInterruptThrown {
				g.RequestInterrupt(mmu.IRQVBlank)

				//throw VBLANK LCD interrupt (if enabled)
				if g.VBlankLCDInterruptEnabled() {
					g.RequestInterrupt(mmu.IRQLCDStat)
				}
				g.vBlankInterruptThrown = true
			}

			//dump output to screen controller over a channel
			result = g.screen
			g.screen = newScreen()
		} else if g.ly > 153 {
			g.vBlankInterruptThrown = false
			g.ly = 0
		}

		//throw coincidence LCD interrupt (if enabled)
		if g.CoincidenceLCDInterruptEnabled() && byte(g.ly) == g.lyc {
			g.stat |= 0x04
			g.RequestInterrupt(mmu.IRQLCDStat)
		}

		//Render scanline
		if g.ly < 144 {
			if g.displayOn {
				if g.bgOn {
					g.RenderBackgroundScanline()
				}

				if g.windowOn {
					g.RenderWindowScanline()
				}

				if g.spritesOn {
					g.RenderSpritesOnScanline()
				}
			}
		}
	}
	return
}

func (g *GPU) CoincidenceLCDInterruptEnabled() bool {
	return (g.Read(STAT) & 0x40) == 0x40
}

func (g *GPU) VBlankLCDInterruptEnabled() bool {
	return (g.Read(STAT) & 0x10) == 0x10
}

func (g *GPU) HblankLCDInterruptEnabled() bool {
	return (g.Read(STAT) & 0x08) == 0x08
}

func (g *GPU) RenderBackgroundScanline() {
	//find where in the tile map we are related to the current scan line + scroll Y (wraps around)
	screenYAdjusted := int(g.ly) + int(g.scrolly)
	initialTilemapOffset := g.bgTilemap + uint16(screenYAdjusted)%256/8*32
	initialLineOffset := uint16(g.scrollx) / 8 % 32

	//find where in the tile we are
	initialTileX := int(g.scrollx) % 8
	initialTileY := screenYAdjusted % 8

	//screen will always draw from X = 0
	g.drawScanline(initialTilemapOffset, initialLineOffset, 0, initialTileX, initialTileY)
}

func (g *GPU) RenderWindowScanline() {
	screenYAdjusted := g.ly - int(g.wy)

	if (g.wx >= 0 && g.wx < 167) && (g.wy >= 0 && g.wy < 144) && screenYAdjusted >= 0 {
		initialTilemapOffset := g.windowTilemap + uint16(screenYAdjusted)/8*32
		screenXAdjusted := int((g.wx - 7) % 255)

		//find where in the tile we are
		initialTileX := screenXAdjusted % 8
		initialTileY := screenYAdjusted % 8

		g.drawScanline(initialTilemapOffset, 0, screenXAdjusted, initialTileX, initialTileY)
	}
}

func (g *GPU) RenderSpritesOnScanline() {
	spriteHeight := 8
	if g.largeSprites {
		spriteHeight = 16
	}

	for _, sprite := range g.oam {
		sy := g.ly - sprite.Y
		if sy < spriteHeight && sy >= 0 {
			tile := sprite.TileId
			if g.largeSprites {
				tile = sprite.TileId & 0xFE
				if sy >= 8 {
					sy -= 8
					if !sprite.FlipV() {
						tile++
					}
				} else if sprite.FlipV() {
					tile++
				}
			}
			g.drawSpriteTileLine(sprite, tile, sy)
		}
	}
}

func (g *GPU) getTilePalIdx(tileID, tileX, tileY int) int {
	addr := uint16(0x8000 | (tileID&0x01FF)<<4 | ((tileY & 0x07) << 1))
	pal0, pal1 := int(g.vram.Read(addr)), int(g.vram.Read(addr+1))
	sx := 1 << (7 - uint(tileX))
	res := 0
	if pal0&sx > 0 {
		res |= 1
	}
	if pal1&sx > 0 {
		res |= 2
	}
	return res
}

func (g *GPU) drawScanline(tilemapOffset, lineOffset uint16, screenX, tileX, tileY int) {
	//get tile to start from
	tileId := g.calculateTileNo(tilemapOffset, lineOffset)

	for ; screenX < DISPLAY_WIDTH; screenX++ {
		//draw the pixel to the screenData data buffer (running through the bgPalette)
		color := g.bgPalette[g.getTilePalIdx(tileId, tileX, tileY)]
		g.screen.Set(screenX, g.ly, color)

		//move along line in tile until you reach the end
		tileX++
		if tileX == 8 {
			tileX = 0
			lineOffset = (lineOffset + 1) % 32

			//get next tile in line
			tileId = g.calculateTileNo(tilemapOffset, lineOffset)
		}
	}
}

func (g *GPU) getTileLine(tileID int, tileY int, flipH, flipV bool) []int {
	result := make([]int, 8)
	if flipV {
		tileY = 7 - tileY
	}

	if flipH {
		for x := 0; x < 8; x++ {
			result[x] = g.getTilePalIdx(tileID, 7-x, tileY)
		}
	} else {
		for x := 0; x < 8; x++ {
			result[x] = g.getTilePalIdx(tileID, x, tileY)
		}
	}
	return result
}

func (g *GPU) noBGAt(x, y int) bool {
	r1, g1, b1, a1 := g.screen.At(x, y).RGBA()
	r2, g2, b2, a2 := g.bgPalette[0].RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func (g *GPU) drawSpriteTileLine(sa *spriteData, tileId, tileY int) {
	y := sa.Y + tileY
	if y < DISPLAY_HEIGHT && y >= 0 {
		for tileX, color := range g.getTileLine(tileId, tileY, sa.FlipH(), sa.FlipV()) {
			if color != 0 {
				x := sa.X + tileX
				if x < DISPLAY_WIDTH && x >= 0 {
					if sa.Priority() || g.noBGAt(x, y) {
						g.screen.Set(x, y, g.objectPalettes[sa.Palette()][color])
					}
				}
			}
		}
	}
}

func (g *GPU) calculateTileNo(tilemapOffset uint16, lineOffset uint16) int {
	tileId := int(g.MMU.Read(uint16(tilemapOffset + lineOffset)))

	//if tile data is 0 then it is signed
	if g.tiledata0 && tileId < 128 {
		tileId += 256
	}

	return tileId
}

func (g *GPU) byteToPalette(b byte) Palette {
	var palette Palette
	palette[0] = gbColors[int(b&0x03)]
	palette[1] = gbColors[int((b>>2)&0x03)]
	palette[2] = gbColors[int((b>>4)&0x03)]
	palette[3] = gbColors[(int(b>>6) & 0x03)]
	return palette
}
