package ppu

type lcdInterrupts byte

const (
	liHBlank lcdInterrupts = 1 << iota
	liVBlank
	liOAM
	liCoincidence
)
