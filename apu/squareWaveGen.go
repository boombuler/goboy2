package apu

type squareWaveGen struct {
	apu      *APU
	ve       *volumeEnvelope
	Volume   float32
	timerCnt int
	dutyIdx  byte
	length   byte

	addr1, addr2, addr3, addr4 uint16

	hi bool

	reg1 byte
	reg2 byte
	reg3 byte
	reg4 byte
}

func newSquareWave(apu *APU, addr1, addr2, addr3, addr4 uint16) *squareWaveGen {
	channel := &squareWaveGen{
		apu:    apu,
		Volume: 1,
		addr1:  addr1,
		addr2:  addr2,
		addr3:  addr3,
		addr4:  addr4,
	}
	channel.ve = newVolumeEnvelop(channel)
	return channel
}

func (s *squareWaveGen) Step(frameStep byte) {
	if frameStep == 7 {
		s.ve.Step()
	}
	useLen := s.useSoundLength()
	if frameStep%2 == 0 && useLen {
		s.length--
	}

	s.timerCnt--

	if s.timerCnt <= 0 {
		s.timerCnt = 2048 - s.freq()
		s.dutyIdx = (s.dutyIdx + 1) % 8

		if !useLen || s.length > 0 {
			if s.duty()&byte(1<<s.dutyIdx) == 0 {
				s.hi = false
			} else {
				s.hi = true
			}
		}
	}

}

func (s *squareWaveGen) CurrentSample() float32 {
	if s.hi {
		return s.Volume * (float32(s.ve.Volume) / 15.0)
	}
	return 0
}

func (s *squareWaveGen) getVolEnvelopCtrl() byte {
	return s.reg2
}

func (s *squareWaveGen) duty() byte {
	// duty setting to wave pattern:
	switch (s.reg1 >> 6) & 0x03 {
	case 0:
		return 0x01
	case 1:
		return 0x81
	case 2:
		return 0x87
	default:
		return 0x7E
	}
}

func (s *squareWaveGen) Read(addr uint16) byte {
	switch addr {
	case s.addr1:
		return s.reg1
	case s.addr2:
		return s.reg2
	case s.addr3:
		return s.reg3
	case s.addr4:
		return s.reg4
	default:
		return 0x00
	}
}

func (s *squareWaveGen) useSoundLength() bool {
	return (s.reg4 & (1 << 6)) != 0
}

func (s *squareWaveGen) setLength() {
	s.length = 64 - (s.reg1 & 0x3F)
}

func (s *squareWaveGen) freq() int {
	v := int(s.reg4 & 7)
	v = v<<7 | int(s.reg3)
	return v
}

func (s *squareWaveGen) setFreq(f int) {
	s.reg3 = byte(f)
	s.reg4 = (s.reg4 & 0xF8) | byte(0x07&(f>>8))
}

func (s *squareWaveGen) Write(addr uint16, val byte) {
	switch addr {
	case s.addr1:
		s.reg1 = val
		s.setLength()
	case s.addr2:
		s.reg2 = val
		s.ve.Reset()
	case s.addr3:
		s.reg3 = val
	case s.addr4:
		s.reg4 = val
		if val&(1<<7) != 0 {
			s.setLength()
		}
	}
}
