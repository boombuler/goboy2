package main

import (
	"os"

	"github.com/boombuler/goboy2/cartridge"
	"github.com/boombuler/goboy2/consts"
	"github.com/boombuler/goboy2/ppu"
)

func newNULLScreen(exitChan <-chan struct{}) chan<- *ppu.ScreenImage {
	screen := make(chan *ppu.ScreenImage)
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

func runMooneyeRom(card *cartridge.Cartridge, compat consts.HardwareCompat) {
	exitChan := make(chan struct{})

	gb := NewGameBoy(card, newNULLScreen(exitChan), compat, exitChan)
	gb.CPU.Dump = *dump
	gb.APU.TestMode = true // no frame limiting, no audio output
	gb.CPU.OnExecOpCode = func(oc string) {
		if oc == "LD B, B" {
			close(exitChan) // Test finished...
		}
	}
	gb.Init(true)

	gb.Run()
	_, _, _, b, c, d, e, _, h, l := gb.CPU.GetRegisterValues()
	if b != 3 || c != 5 || d != 8 || e != 13 || h != 21 || l != 34 {
		os.Exit(1)

	} else {
		os.Exit(0)
	}
}
