package main

import (
	"github.com/boombuler/goboy2/apu"
	"github.com/boombuler/goboy2/cartridge"
	"github.com/boombuler/goboy2/consts"
	"github.com/boombuler/goboy2/cpu"
	"github.com/boombuler/goboy2/input"
	"github.com/boombuler/goboy2/mmu"
	"github.com/boombuler/goboy2/ppu"
	"github.com/boombuler/goboy2/serial"
	"github.com/boombuler/goboy2/timer"
)

const compatAuto consts.HardwareCompat = -1

type GameBoy struct {
	exitChan <-chan struct{}
	compat   consts.HardwareCompat
	MMU      mmu.MMU
	CPU      *cpu.CPU
	PPU      *ppu.PPU
	APU      *apu.APU
	Timer    *timer.Timer
	Input    *input.Keyboard
	Serial   *serial.Serial
}

// NewGameBoy creates a new gameboy for the given cartridge
func NewGameBoy(c *cartridge.Cartridge, screen chan<- *ppu.ScreenImage, hw consts.HardwareCompat, exitChan <-chan struct{}) *GameBoy {
	gb := new(GameBoy)
	gb.exitChan = exitChan

	if hw == compatAuto {
		if c.GBC {
			hw = consts.GBC
		} else {
			hw = consts.DMG
		}
	}

	gb.compat = hw
	gb.MMU = mmu.New(hw)
	gb.APU = apu.New(gb.MMU)
	gb.CPU = cpu.New(gb.MMU, hw)
	gb.PPU = ppu.New(gb.MMU, hw, screen, exitChan)
	gb.Timer = timer.New(gb.MMU, hw)
	gb.Serial = serial.New(gb.MMU)
	gb.Input = input.NewKeyboard(gb.MMU)
	gb.MMU.LoadCartridge(c)
	return gb
}

// Run starts the emulation until the exit chan is closed.
func (gb *GameBoy) Run() {
	if err := gb.APU.Start(); err != nil {
		panic(err)
	}
	dsTick := false
	for {
		select {
		case _, _ = <-gb.exitChan:
			gb.APU.Stop()
			return
		default:
			gb.Timer.Prepare()
			gb.CPU.Step()
			gb.MMU.Step()
			gb.Timer.Step()
			gb.Serial.Step()
			if !dsTick {
				gb.APU.Step()
				gb.PPU.Step()
				dsTick = gb.CPU.DoubleSpeed()
			} else {
				dsTick = false
			}
		}
	}
}

// InitNoBOOT brings the gameboy to the state after the bootrom finished
func (gb *GameBoy) Init(noBoot bool) {
	gb.MMU.Init(noBoot)
	gb.CPU.Init(noBoot)
	gb.Timer.Init(noBoot)
	gb.APU.Init(noBoot)
	gb.PPU.Init(noBoot)
}
