package ppu

type lcdInterrupts byte

const (
	liHBlank lcdInterrupts = 1 << (3 + iota)
	liVBlank
	liOAM
	liCoincidence

	liALL = liHBlank | liVBlank | liOAM | liCoincidence
)

type ppuPhase interface {
	state() ppuState
	step(ppu *PPU) bool
	start(ppu *PPU) bool
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
