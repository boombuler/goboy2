package timer

import (
	"github.com/boombuler/goboy2/consts"
	"github.com/boombuler/goboy2/mmu"
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
	hw   consts.HardwareCompat
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
func New(mmu mmu.MMU, hw consts.HardwareCompat) *Timer {
	t := new(Timer)
	t.mmu = mmu
	t.hw = hw
	mmu.AddIODevice(t, consts.AddrDivider, consts.AddrTIMA, consts.AddrModulo, consts.AddrCtrl)
	return t
}

func (t *Timer) Init(noBoot bool) {
	// I guess the timer runs before the cpu boots up.
	if t.hw == consts.GBC {
		t.div = 0x8970
		if noBoot {
			t.div += 0x9D05
		}
	} else {
		t.div = 0x0245
		if noBoot {
			t.div += 0xA985
		}
	}
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
	case consts.AddrDivider:
		return byte(t.div >> 8)
	case consts.AddrTIMA:
		return t.tima
	case consts.AddrModulo:
		return t.tma
	case consts.AddrCtrl:
		return t.tac | 0xF8
	}
	return 0x00
}

// Write to the timer IO registers
func (t *Timer) Write(addr uint16, value byte) {
	switch addr {
	case consts.AddrDivider:
		t.setDiv(0)
	case consts.AddrTIMA:
		if t.overflow < osInterrupt {
			t.tima = value
			t.overflow = osNone
		}
	case consts.AddrModulo:
		t.tma = value
	case consts.AddrCtrl:
		oldHi := t.hi()
		t.tac = value & 0x07
		if oldHi && !t.hi() { // Check falling edge
			t.incTIMA()
		}
	}
}
