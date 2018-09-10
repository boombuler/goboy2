package apu

const (
	AddrNR21 uint16 = 0xFF16
	AddrNR22 uint16 = 0xFF17
	AddrNR23 uint16 = 0xFF18
	AddrNR24 uint16 = 0xFF19
)

type soundChannel2 struct {
	apu      *APU
	ve       *volumeEnvelope
	Volume   float32
	timerCnt int
	dutyIdx  byte
	length   byte

	sample float32

	nr21 byte
	nr22 byte
	nr23 byte
	nr24 byte
}

func newSC2(apu *APU) *soundChannel2 {
	channel := &soundChannel2{
		apu:    apu,
		Volume: 1,
	}
	channel.ve = newVolumeEnvelop(channel)
	return channel
}

func (s *soundChannel2) Step() {
	s.ve.Step()
	s.timerCnt--
	if s.timerCnt <= 0 {
		s.timerCnt = s.freq()
		s.dutyIdx = (s.dutyIdx + 1) % 8

		if useLen := !s.useSoundLength(); useLen || s.length > 0 {
			if s.duty()&byte(1<<s.dutyIdx) == 0 {
				s.sample = 0
			} else {
				s.sample = s.Volume * (float32(s.ve.Volume) / 15.0)
			}

			if useLen {
				s.length--
			}
		}
	}

}

func (s *soundChannel2) CurrentSample() float32 {
	return s.sample
}

func (s *soundChannel2) getVolEnvelopCtrl() byte {
	return s.nr22
}

func (s *soundChannel2) duty() byte {
	// duty setting to wave pattern:
	switch (s.nr21 >> 6) & 0x03 {
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

func (s *soundChannel2) Read(addr uint16) byte {
	switch addr {
	case AddrNR21:
		return s.nr21
	case AddrNR22:
		return s.nr22
	case AddrNR23:
		return s.nr23
	case AddrNR24:
		return s.nr24
	default:
		return 0x00
	}
}

func (s *soundChannel2) useSoundLength() bool {
	return (s.nr24 & (1 << 6)) != 0
}

func (s *soundChannel2) setLength() {
	s.length = 64 - (s.nr21 & 0x3F)
}

func (s *soundChannel2) freq() int {
	v := int(s.nr24 & 7)
	v = v<<7 | int(s.nr23)
	return gbTicksPerSecond / (2048 - v)
}

func (s *soundChannel2) Write(addr uint16, val byte) {
	switch addr {
	case AddrNR21:
		s.nr21 = val
		s.setLength()
	case AddrNR22:
		s.nr22 = val
		s.ve.Reset()
	case AddrNR23:
		s.nr23 = val
	case AddrNR24:
		s.nr24 = val
		if val&(1<<7) != 0 {
			s.setLength()
		}
	}
}
