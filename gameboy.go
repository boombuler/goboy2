package main

import (
	"goboy2/apu"
	"goboy2/cartridge"
	"goboy2/cpu"
	"goboy2/gpu"
	"goboy2/input"
	"goboy2/mmu"
	"goboy2/timer"
	"image"
	"time"
)

type GameBoy struct {
	MMU   mmu.MMU
	CPU   *cpu.CPU
	gpu   *gpu.GPU
	APU   *apu.APU
	timer *timer.Timer
	kb    *input.Keyboard
}

func NewGameBoy(c *cartridge.Cartridge) *GameBoy {
	gb := new(GameBoy)
	gb.MMU = mmu.New()
	gb.APU = apu.New(gb.MMU)
	gb.CPU = cpu.New(gb.MMU)
	gb.gpu = gpu.New(gb.MMU)
	gb.timer = timer.New(gb.MMU)
	gb.kb = input.NewKeyboard(gb.MMU)
	gb.MMU.LoadCartridge(c)
	return gb
}

const frameduration = (time.Second / 4194304) * 70224

// StepFrame keeps the gameboy running until the next frame is done
func (gb *GameBoy) StepFrame() *image.RGBA {
	starttime := time.Now()
	for {
		gb.timer.Prepare()
		gb.CPU.Step()
		gb.MMU.Step()
		gb.timer.Step()
		gb.APU.Step()

		if img := gb.gpu.Step(4); img != nil {
			sleepTime := frameduration - time.Now().Sub(starttime)
			time.Sleep(sleepTime)
			return img
		}
	}
}

func (gb *GameBoy) SetupNoBootRom() {
	gb.MMU.Write(mmu.AddrBootmodeFlag, 0x01)
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
