package ppu

import (
	"goboy2/consts"
	"image"
	"image/color"
	"sync"
)

var emptyScreen = newScreen()

var screenPool = &sync.Pool{
	New: func() interface{} {
		return image.NewRGBA(image.Rect(0, 0, consts.DisplayWidth, consts.DisplayHeight))
	},
}

func FreeScreen(img *image.RGBA) {
	screenPool.Put(img)
}

func dropFrames(output chan<- *image.RGBA) chan<- *image.RGBA {
	input := make(chan *image.RGBA)

	go func() {
		lastImg := <-input
		for {
			select {
			case img := <-input:
				FreeScreen(lastImg)
				lastImg = img
			case output <- lastImg:
			}
		}
	}()

	return input
}

func newScreen() *image.RGBA {
	return screenPool.Get().(*image.RGBA)
}

type palette interface {
	toColor(pIdx int, val byte) color.Color
}

type gbPalette uint16

type gbcPalette struct {
	IndexAdr uint16
	idx      int
	autoInc  bool
	data     [32]color.RGBA
}

func newGBCPalette(idxAddr uint16) *gbcPalette {
	r := new(gbcPalette)
	r.IndexAdr = idxAddr
	for i := 0; i < 32; i++ {
		r.data[i].A = 0xFF
		r.data[i].R = 0xF8
		r.data[i].G = 0xF8
		r.data[i].B = 0xF8
	}
	return r
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
	col := p.data[p.idx>>1]

	if hi {
		G := (col.G >> 3) & 0x1F
		B := (col.B >> 3) & 0x1F
		return G>>3 | B<<2
	}

	R := (col.R >> 3) & 0x1F
	G := (col.G >> 3) & 0x1F
	return R | (G & 0x07 << 5)
}

func (p *gbcPalette) Write(addr uint16, val byte) {
	if addr == p.IndexAdr {
		p.idx = int(val & 0x3f)
		p.autoInc = (val & (1 << 7)) != 0
	} else {
		hi := p.idx&1 == 1
		idx := p.idx >> 1
		if hi {
			gVal := (val & 0x03) << 6
			bVal := (val & 0x7C) << 1
			p.data[idx].B = bVal
			p.data[idx].G = gVal | (p.data[idx].G & 0x38)
		}
		rVal := (val & 0x1F) << 3
		gVal := (val & 0xE0) >> 2
		p.data[idx].R = rVal
		p.data[idx].G = gVal | (p.data[idx].G & 0xC0)

		if p.autoInc {
			p.idx++
		}
	}
}

func (p *gbcPalette) toColor(pIdx int, val byte) color.Color {
	val &= 0x03
	idx := (pIdx << 2) | int(val)

	return p.data[idx]
}

var gbColors = []color.Color{
	color.RGBAModel.Convert(color.Gray{0xEB}),
	color.RGBAModel.Convert(color.Gray{0xC4}),
	color.RGBAModel.Convert(color.Gray{0x60}),
	color.RGBAModel.Convert(color.Gray{0x00}),
}

func (p gbPalette) toColor(pIdx int, val byte) color.Color {
	pal := byte(p)
	if pIdx == 1 {
		pal = byte(p >> 8)
	}

	shift := (val & 0x03) * 2
	color := (pal >> shift)
	return gbColors[0x03&color]
}
