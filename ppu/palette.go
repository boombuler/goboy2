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
	data     [64]byte
}

func (p *gbcPalette) Read(addr uint16) byte {
	if addr == p.IndexAdr {
		inc := byte(0x00)
		if p.autoInc {
			inc = 0x80
		}
		return byte(p.idx) | inc | 0x40
	} else {
		return p.data[p.idx]
	}
}

func (p *gbcPalette) Write(addr uint16, val byte) {
	if addr == p.IndexAdr {
		p.idx = int(val & 0x3f)
		p.autoInc = (val & (1 << 7)) != 0
	} else {
		p.data[p.idx] = val
		if p.autoInc {
			p.idx++
		}
	}
}

func (p *gbcPalette) toColor(pIdx int, val byte) color.Color {
	val &= 0x03
	idx := (pIdx * 8) + int(val*2)

	colorV := uint16(p.data[idx+1])<<8 | uint16(p.data[idx])
	red := byte(colorV & 0x1F)
	green := byte((colorV >> 5) & 0x1F)
	blue := byte((colorV >> 10) & 0x1F)

	return color.RGBA{
		R: red << 3,
		G: green << 3,
		B: blue << 3,
		A: 0xFF,
	}
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
