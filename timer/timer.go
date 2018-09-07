package timer

import (
	"goboy2/mmu"
)

const (
	// AddrDivider is the address of the timer DIV register
	AddrDivider = 0xFF04
	// AddrTIMA is the address of the timer TIMA register
	AddrTIMA = 0xFF05
	// AddrModulo is the address of the timer TMA register
	AddrModulo = 0xFF06
	// AddrCtrl is the address of the timer TAC register
	AddrCtrl = 0xFF07
)

var clockSpeedBit = [4]uint16{1 << 9, 1 << 3, 1 << 5, 1 << 7}

const (
	// bit from the tac register to enable timer
	enabled = 0x04
	// speed mask for tac register
	speed = 0x03
)

type overflowState byte

// Timer encapsulates the Gameboy hardware timer
type Timer struct {
	mmu  mmu.MMU
	div  uint16
	tac  byte
	tima byte
	tma  byte

	overflow overflowState
}

const (
	osNone overflowState = iota
	osOverflow
	osInterrupt
)

// New creates a new timer for the given mmu and connects it to the given clock
func New(mmu mmu.MMU) *Timer {
	t := new(Timer)
	t.mmu = mmu
	mmu.AddIODevice(t, AddrDivider, AddrTIMA, AddrModulo, AddrCtrl)
	return t
}

// Prepare the timer for the next cpu tick.
func (t *Timer) Prepare() {
	t.setDiv(t.div + 4)
	if t.overflow == osInterrupt {
		t.tima = t.tma
		t.mmu.RequestInterrupt(mmu.IRQTimer)
	}
}

// Step the timer after the cpu is done
func (t *Timer) Step() {
	if t.overflow == osOverflow {
		t.overflow = osInterrupt
	} else if t.overflow == osInterrupt {
		t.tima = t.tma
		t.overflow = osNone
	}
}

// check the timer bit + enabled for the falling edge detector
func (t *Timer) hi() bool {
	return (t.tac&enabled != 0) &&
		(t.div&clockSpeedBit[t.tac&speed]) != 0
}

// set internal counter and check for falling edge
func (t *Timer) setDiv(val uint16) {
	oldHi := t.hi()
	t.div = val
	if oldHi && !t.hi() {
		t.incTIMA()
	}
}

// increment TIMA register, and maybe start overflowing.
func (t *Timer) incTIMA() {
	t.tima++
	if t.tima == 0 {
		t.overflow = osOverflow
	}
}

// Read from the timer IO registers
func (t *Timer) Read(addr uint16) byte {
	switch addr {
	case AddrDivider:
		return byte(t.div >> 8)
	case AddrTIMA:
		return t.tima
	case AddrModulo:
		return t.tma
	case AddrCtrl:
		return t.tac
	}
	return 0x00
}

// Write to the timer IO registers
func (t *Timer) Write(addr uint16, value byte) {
	switch addr {
	case AddrDivider:
		t.setDiv(0)
	case AddrTIMA:
		if t.overflow < osInterrupt {
			t.tima = value
			t.overflow = osNone
		}
	case AddrModulo:
		t.tma = value
	case AddrCtrl:
		oldHi := t.hi()
		t.tac = value
		if oldHi && !t.hi() { // Check falling edge
			t.incTIMA()
		}
	}
}
