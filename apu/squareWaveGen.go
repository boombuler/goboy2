package apu

type squareWaveGen struct {
	apu           *APU
	ve            *volumeEnvelope
	timerCnt      int
	dutyIdx       byte
	lengthCounter byte

	addr1, addr2, addr3, addr4 uint16

	hi bool

	dacEnabled bool
	dutyMode   byte
	lengthLoad byte
	timerLoad  int
	useLength  bool
}

func newSquareWave(apu *APU, addr1, addr2, addr3, addr4 uint16) *squareWaveGen {
	channel := &squareWaveGen{
		apu:   apu,
		addr1: addr1,
		addr2: addr2,
		addr3: addr3,
		addr4: addr4,
		ve:    new(volumeEnvelope),
	}
	return channel
}

func (s *squareWaveGen) Step(frameStep byte) {
	if frameStep == 7 {
		s.ve.Step()
	}

	if frameStep%2 == 0 && s.useLength {
		s.lengthCounter--
	}

	s.timerCnt -= 2

	if s.timerCnt <= 0 {
		s.timerCnt = 2048 - s.timerLoad
		s.dutyIdx = (s.dutyIdx + 1) % 8

		if !s.useLength || s.lengthCounter > 0 {
			if s.duty()&byte(1<<s.dutyIdx) == 0 {
				s.hi = false
			} else {
				s.hi = true
			}
		}
	}
}

func (s *squareWaveGen) CurrentSample() float32 {
	if s.hi && s.dacEnabled {
		return s.ve.Volume()
	}
	return 0
}

func (s *squareWaveGen) duty() byte {
	// duty setting to wave pattern:
	switch s.dutyMode {
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

func (s *squareWaveGen) setLength() {
	s.lengthCounter = 64 - s.lengthLoad
}

func (s *squareWaveGen) trigger() {
	if s.lengthCounter == 0 {
		s.lengthCounter = 64
	}
	s.timerCnt = 2048 - s.timerLoad
	s.ve.Reset()
}

func (s *squareWaveGen) Read(addr uint16) byte {
	switch addr {
	case s.addr1:
		return (s.lengthLoad & 0x3F) | ((s.dutyMode & 0x03) << 6)
	case s.addr2:
		var inc byte
		if s.ve.Increase {
			inc = 1
		}
		return byte(s.ve.periodLoad&0x07) | (inc << 3) | (s.ve.VolumeLoad << 4)
	case s.addr3:
		return byte(s.timerLoad)
	case s.addr4:
		var useLen byte
		if s.useLength {
			useLen = 1
		}

		return byte((s.timerLoad>>8)&0x07) | (useLen << 6)
	default:
		return 0x00
	}
}

func (s *squareWaveGen) Write(addr uint16, val byte) {
	switch addr {
	case s.addr1:
		s.lengthLoad = val & 0x3F
		s.dutyMode = (val >> 6) & 0x03
	case s.addr2:
		s.dacEnabled = val&0xF8 != 0
		s.ve.VolumeLoad = (val >> 4) & 0x0F
		s.ve.Increase = (val & 0x08) != 0
		s.ve.periodLoad = (val & 0x07)
		s.ve.Reset()
	case s.addr3:
		s.timerLoad = (s.timerLoad & 0x0700) | int(val)
	case s.addr4:
		s.timerLoad = (s.timerLoad & 0xFF) | (int(val&0x07) << 8)
		s.useLength = val&0x40 != 0
		if val&(1<<7) != 0 {
			s.trigger()
		}
	}
}
