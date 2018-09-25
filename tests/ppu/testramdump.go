package main

import (
	"goboy2/mmu"
	"goboy2/ppu"
	"image"
	"image/png"
	"os"
)

func main() {
	mmu := mmu.New()

	screen := make(chan *image.RGBA)
	exit := make(chan struct{})

	dump, _ := os.Open("vram.mem")
	defer dump.Close()

	ppu := ppu.New(mmu, screen)
	ppu.LoadDump(dump)

	go func() {
		for {
			select {
			case _, _ = <-exit:
				return
			default:
				ppu.Step()
			}
		}
	}()

	var img *image.RGBA
	for i := 0; i < 5; i++ {
		img = <-screen
	}
	close(exit)
	f, _ := os.Create("render.png")
	defer f.Close()
	png.Encode(f, img)

}
