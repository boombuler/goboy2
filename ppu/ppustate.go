package ppu

type ppuPhase interface {
	state() ppuState
	step(ppu *PPU) interface{}
	start(ppu *PPU)
}

type ppuState byte

const (
	sHBlank        ppuState = 0x00
	sVBlank        ppuState = 0x01
	sOAMRead       ppuState = 0x02
	sPixelTransfer ppuState = 0x03
)

func (s ppuState) canAccessOAM() bool {
	switch s {
	case sOAMRead, sPixelTransfer:
		return false
	default:
		return true
	}
}

func (s ppuState) canAccessVRAM() bool {
	if s == sPixelTransfer {
		return false
	}
	return true
}
