package serial

import "goboy2/mmu"

type Serial struct {
	mmu      mmu.MMU
	transfer SerialTransfer

	sb                 byte
	sc                 byte
	transferInProgress bool
	divider            int
}

const (
	addrSB = 0xFF01 // Serial Transfer Data
	addrSC = 0xFF02 // Serial Transfer Control

	gbFreq           = 4194304
	gbTicksPerSecond = gbFreq / 4 // use M-Cylces
	serialTickDiv    = gbTicksPerSecond / 2048
)

type SerialTransfer interface {
	Exchange(val byte) byte
}

type nullTransfer struct{}

func (n nullTransfer) Exchange(val byte) byte {
	return 0
}

func New(mmu mmu.MMU) *Serial {
	res := &Serial{
		mmu:      mmu,
		transfer: nullTransfer{},
	}
	mmu.AddIODevice(res, addrSB, addrSC)
	return res
}

func (s *Serial) Read(addr uint16) byte {
	switch addr {
	case addrSB:
		return s.sb
	case addrSC:
		return s.sc | 0x7E
	default:
		return 0xFF
	}
}

func (s *Serial) startTransfer() {
	s.transferInProgress = true
	s.divider = 0
}

func (s *Serial) Write(addr uint16, val byte) {
	switch addr {
	case addrSB:
		s.sb = val
	case addrSC:
		s.sc = val
		if (s.sc & 0x80) != 0 {
			s.startTransfer()
		}
	}
}

func (s *Serial) Step() {
	if s.transferInProgress {
		if s.divider++; s.divider >= serialTickDiv {
			s.transferInProgress = false
			s.sb = s.transfer.Exchange(s.sb)
			s.mmu.RequestInterrupt(mmu.IRQSerial)
		}
	}
}
