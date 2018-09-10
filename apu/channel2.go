package apu

const (
	AddrNR21 uint16 = 0xFF16
	AddrNR22 uint16 = 0xFF17
	AddrNR23 uint16 = 0xFF18
	AddrNR24 uint16 = 0xFF19
)

type soundChannel2 struct {
	apu *APU

	nr21 byte
	nr22 byte
	nr23 byte
	nr24 byte
}

func (s *soundChannel2) Step() {

}

func (s *soundChannel2) GenerateSamples(buffer []int16) bool {
	return false
}

func (s *soundChannel2) getVolEnvelopCtrl() byte {
	return s.nr22
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

func (s *soundChannel2) Write(addr uint16, val byte) {
	switch addr {
	case AddrNR21:
		s.nr21 = val
	case AddrNR22:
		s.nr22 = val
	case AddrNR23:
		s.nr23 = val
	case AddrNR24:
		s.nr24 = val
	}
}
