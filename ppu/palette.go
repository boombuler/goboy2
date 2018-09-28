package ppu

import (
	"fmt"
	"sync"
)

var emptyScreen = newScreen()

const colorShift = 3 // Amount to shift the gameboy color to the right, for RGB values...

var screenPool = &sync.Pool{
	New: func() interface{} {
		return new(ScreenImage)
	},
}

func FreeScreen(img *ScreenImage) {
	screenPool.Put(img)
}

func dropFrames(output chan<- *ScreenImage, exitChan <-chan struct{}) chan<- *ScreenImage {
	input := make(chan *ScreenImage)

	go func() {
		lastImg := <-input
		for {
			select {
			case _, _ = <-exitChan:
				return
			case img := <-input:
				FreeScreen(lastImg)
				lastImg = img
			case output <- lastImg:
			}
		}
	}()

	return input
}

func newScreen() *ScreenImage {
	return screenPool.Get().(*ScreenImage)
}

type palette interface {
	toColor(pIdx int, val byte) RGB
}

type gbPalette uint16

type gbcPalette struct {
	IndexAdr uint16
	idx      int
	autoInc  bool
	data     [32]RGB
}

func newGBCPalette(idxAddr uint16) *gbcPalette {
	r := new(gbcPalette)
	r.IndexAdr = idxAddr
	for i := 0; i < 32; i++ {
		r.data[i].R = 0xF8
		r.data[i].G = 0xF8
		r.data[i].B = 0xF8
	}
	return r
}

func (p *gbcPalette) PrintRam() {
	for i, c := range p.data {
		hi, lo := getColorBytes(c)
		if i%4 == 0 {
			fmt.Printf("%d: ", i/4)
		}

		if i%4 == 3 {
			fmt.Printf("%02X%02X\n", hi, lo)
		} else {
			fmt.Printf("%02X%02X ", hi, lo)
		}
	}
}

func getColorBytes(col RGB) (hi byte, lo byte) {
	R := (col.R >> colorShift) & 0x1F
	G := (col.G >> colorShift) & 0x1F
	B := (col.B >> colorShift) & 0x1F
	hi = G>>3 | B<<2
	lo = R | (G & 0x07 << 5)
	return
}

func setColorBytes(hi, lo byte) RGB {
	r := lo & 0x1F
	g := ((lo & 0xE0) >> 5) | ((hi & 0x03) << 3)
	b := (hi >> 2) & 0x1F
	return RGB{
		R: r << colorShift,
		G: g << colorShift,
		B: b << colorShift,
	}
}

func (p *gbcPalette) Read(addr uint16) byte {
	if addr == p.IndexAdr {
		inc := byte(0x00)
		if p.autoInc {
			inc = 0x80
		}
		return byte(p.idx) | inc | 0x40
	}

	hi := p.idx&1 != 0
	cHi, cLo := getColorBytes(p.data[p.idx>>1])
	if hi {
		return cHi
	}
	return cLo
}

func (p *gbcPalette) Write(addr uint16, val byte) {
	if addr == p.IndexAdr {
		p.idx = int(val & 0x3f)
		p.autoInc = (val & (1 << 7)) != 0
	} else {
		hi := p.idx&1 == 1
		idx := p.idx >> 1
		pHi, pLo := getColorBytes(p.data[idx])
		if hi {
			p.data[idx] = setColorBytes(val, pLo)
		} else {
			p.data[idx] = setColorBytes(pHi, val)
		}

		if p.autoInc {
			p.idx++
		}
	}
}

func (p *gbcPalette) toColor(pIdx int, val byte) RGB {
	return p.data[(pIdx<<2)|int(val&0x03)]
}

var gbColors = []RGB{
	RGB{0xEB, 0xEB, 0xEB},
	RGB{0xC4, 0xC4, 0xC4},
	RGB{0x60, 0x60, 0x60},
	RGB{0x00, 0x00, 0x00},
}

func (p *gbPalette) toColor(pIdx int, val byte) RGB {
	pal := byte(*p)
	if pIdx == 1 {
		pal = byte(*p >> 8)
	}

	shift := (val & 0x03) * 2
	color := (pal >> shift)
	return gbColors[0x03&color]
}
