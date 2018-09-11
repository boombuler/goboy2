package apu

const (
	AddrNR10 uint16 = 0xFF10
	AddrNR11 uint16 = 0xFF11
	AddrNR12 uint16 = 0xFF12
	AddrNR13 uint16 = 0xFF13
	AddrNR14 uint16 = 0xFF14
)

type soundChannel1 struct {
	*squareWaveGen
	sweepCtrl  byte
	timer      int
	overflowed bool
}

func newSC1(apu *APU) *soundChannel1 {
	return &soundChannel1{
		newSquareWave(apu, AddrNR11, AddrNR12, AddrNR13, AddrNR14),
		0, 0, true,
	}
}

func (s *soundChannel1) Step(step byte) {
	if (step == 2 || step == 6) && s.sweepTime() > 0 {
		if s.timer--; s.timer <= 0 {
			s.timer = s.sweepTime()
			// sweep
			if s.sweepShift() > 0 {
				curFreq := s.squareWaveGen.freq()
				amount := curFreq >> s.sweepShift()
				if s.sweepUp() {
					curFreq = (curFreq - amount)
				} else {
					curFreq = (curFreq + amount)
				}

				if curFreq > 2047 {
					s.overflowed = true
				} else {
					s.squareWaveGen.setFreq(curFreq)
				}
			}
		}
	}
	s.squareWaveGen.Step(step)
}

func (s *soundChannel1) sweepTime() int {
	return int(s.sweepCtrl>>4) & 0x03
}

func (s *soundChannel1) sweepUp() bool {
	return (s.sweepCtrl & 0x08) != 0
}
func (s *soundChannel1) sweepShift() byte {
	return s.sweepCtrl & 0x07
}

func (s *soundChannel1) Read(addr uint16) byte {
	if addr == AddrNR10 {
		return s.sweepCtrl
	}
	return s.squareWaveGen.Read(addr)
}

func (s *soundChannel1) Write(addr uint16, val byte) {
	if addr == AddrNR10 {
		s.sweepCtrl = val
		s.overflowed = false
		return
	}
	s.squareWaveGen.Write(addr, val)
}
