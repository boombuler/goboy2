package apu

type sweepSquareWaveGen struct {
	*squareWaveGen
	sweepCtrl  byte
	timer      int
	overflowed bool
}

func newSweepSquareWaveGen(apu *APU) *sweepSquareWaveGen {
	return &sweepSquareWaveGen{
		newSquareWave(apu, addrNR11, addrNR12, addrNR13, addrNR14),
		0, 0, true,
	}
}

func (s *sweepSquareWaveGen) Reset() {
	s.squareWaveGen.Reset()
	s.sweepCtrl = 0
	s.timer = 0
	s.overflowed = true
}

func (s *sweepSquareWaveGen) Step(step sequencerStep) {
	if (step&ssSweep == ssSweep) && s.sweepTime() > 0 {
		if s.timer--; s.timer <= 0 {
			s.timer = s.sweepTime()
			// sweep
			if s.sweepShift() > 0 {
				curFreq := s.squareWaveGen.timerLoad
				amount := curFreq >> s.sweepShift()
				if s.sweepUp() {
					curFreq = (curFreq - amount)
				} else {
					curFreq = (curFreq + amount)
				}

				if curFreq > 2047 {
					s.overflowed = true
				} else {
					s.squareWaveGen.timerLoad = curFreq
				}
			}
		}
	}
	s.squareWaveGen.Step(step)
}

func (s *sweepSquareWaveGen) sweepTime() int {
	return int(s.sweepCtrl>>4) & 0x03
}

func (s *sweepSquareWaveGen) sweepUp() bool {
	return (s.sweepCtrl & 0x08) != 0
}
func (s *sweepSquareWaveGen) sweepShift() byte {
	return s.sweepCtrl & 0x07
}

func (s *sweepSquareWaveGen) Read(addr uint16) byte {
	if addr == addrNR10 {
		return s.sweepCtrl | 0x80
	}
	return s.squareWaveGen.Read(addr)
}

func (s *sweepSquareWaveGen) Write(addr uint16, val byte) {
	if addr == addrNR10 {
		s.sweepCtrl = val
		s.overflowed = false
		return
	}
	s.squareWaveGen.Write(addr, val)
}
