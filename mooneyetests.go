package main

import (
	"fmt"
	"goboy2/cartridge"
	"image"
)

func newNULLScreen(exitChan <-chan struct{}) chan<- *image.RGBA {
	screen := make(chan *image.RGBA)
	go func() {
		for {
			select {
			case _, _ = <-screen:
				break
			case _, _ = <-exitChan:
				return
			}
		}
	}()
	return screen
}

func runMooneyeRom(card *cartridge.Cartridge) {

	exitChan := make(chan struct{})

	gb := NewGameBoy(card, newNULLScreen(exitChan))
	gb.apu.TestMode = true // no frame limiting, no audio output
	gb.CPU.OnExecOpCode = func(oc string) {
		if oc == "LD B, B" {
			close(exitChan) // Test finished...
		}
	}

	gb.Run(exitChan)
	_, _, _, b, c, d, e, _, h, l := gb.CPU.GetRegisterValues()
	if b != 3 || c != 5 || d != 8 || e != 13 || h != 21 || l != 34 {
		fmt.Println("\033[0;31m FAILED\033[0;37m")
	} else {
		fmt.Println("\033[0;32m OK\033[0;37m")
	}
}