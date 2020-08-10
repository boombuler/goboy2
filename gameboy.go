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

type GameBoy struct {
	exitChan <-chan struct{}
	MMU      mmu.MMU
	CPU      *cpu.CPU
	ppu      *ppu.PPU
	apu      *apu.APU
	timer    *timer.Timer
	kb       *input.Keyboard
	serial   *serial.Serial
}

// NewGameBoy creates a new gameboy for the given cartridge
func NewGameBoy(c *cartridge.Cartridge, screen chan<- *ppu.ScreenImage, forceGBC bool, exitChan <-chan struct{}) *GameBoy {
	gb := new(GameBoy)
	gb.exitChan = exitChan
	gb.MMU = mmu.New(c.GBC || forceGBC)
	gb.apu = apu.New(gb.MMU)
	gb.CPU = cpu.New(gb.MMU)
	gb.ppu = ppu.New(gb.MMU, screen, exitChan)
	gb.timer = timer.New(gb.MMU)
	gb.serial = serial.New(gb.MMU)
	gb.kb = input.NewKeyboard(gb.MMU)
	gb.MMU.LoadCartridge(c)
	return gb
}

// Run starts the emulation until the exit chan is closed.
func (gb *GameBoy) Run() {
	if err := gb.apu.Start(); err != nil {
		panic(err)
	}
	dsTick := false
	for {
		select {
		case _, _ = <-gb.exitChan:
			gb.apu.Stop()
			return
		default:
			gb.timer.Prepare()
			gb.CPU.Step()
			gb.MMU.Step()
			gb.timer.Step()
			gb.serial.Step()
			if !dsTick {
				gb.apu.Step()
				gb.ppu.Step()
				dsTick = gb.CPU.DoubleSpeed()
			} else {
				dsTick = false
			}
		}
	}
}

// SetupNoBootRom brings the gameboy to the state after the bootrom finished
func (gb *GameBoy) SetupNoBootRom() {
	gb.MMU.Write(consts.AddrBootmodeFlag, 0x01)
	gb.CPU.SetRegisterValues(0x0100, 0xFFFE, 0x01, 0x00, 0x13, 0x00, 0xD8, 0xB0, 0x01, 0x4D)
	gb.MMU.Write(0xFF05, 0x00)
	gb.MMU.Write(0xFF06, 0x00)
	gb.MMU.Write(0xFF07, 0x00)
	gb.MMU.Write(0xFF10, 0x80)
	gb.MMU.Write(0xFF11, 0xBF)
	gb.MMU.Write(0xFF12, 0xF3)
	gb.MMU.Write(0xFF14, 0xBF)
	gb.MMU.Write(0xFF16, 0x3F)
	gb.MMU.Write(0xFF17, 0x00)
	gb.MMU.Write(0xFF19, 0xBF)
	gb.MMU.Write(0xFF1A, 0x7F)
	gb.MMU.Write(0xFF1B, 0xFF)
	gb.MMU.Write(0xFF1C, 0x9F)
	gb.MMU.Write(0xFF1E, 0xBF)
	gb.MMU.Write(0xFF20, 0xFF)
	gb.MMU.Write(0xFF21, 0x00)
	gb.MMU.Write(0xFF22, 0x00)
	gb.MMU.Write(0xFF23, 0xBF)
	gb.MMU.Write(0xFF24, 0x77)
	gb.MMU.Write(0xFF25, 0xF3)
	gb.MMU.Write(0xFF26, 0xF1)
	gb.MMU.Write(0xFF40, 0x91)
	gb.MMU.Write(0xFF42, 0x00)
	gb.MMU.Write(0xFF43, 0x00)
	gb.MMU.Write(0xFF45, 0x00)
	gb.MMU.Write(0xFF47, 0xFC)
	gb.MMU.Write(0xFF48, 0xFF)
	gb.MMU.Write(0xFF49, 0xFF)
	gb.MMU.Write(0xFF4A, 0x00)
	gb.MMU.Write(0xFF4B, 0x00)
	gb.MMU.Write(0xFF50, 0x00)
	gb.MMU.Write(0xFFFF, 0x00)
}
