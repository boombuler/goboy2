package apu

/*
       Noise
NR41 --LL LLLL Length load (64-L)
NR42 VVVV APPP Starting volume, Envelope add mode, period
NR43 SSSS WDDD Clock shift, Width mode of LFSR, Divisor code
NR44 TL-- ---- Trigger, Length enable
*/
var (
	divisors = []byte{8, 16, 32, 48, 64, 80, 96, 112}
)

type noiseGen struct {
	ve *volumeEnvelope

	hi   bool
	lfsr uint16

	timerCnt   int
	length     byte
	lengthLoad byte
	useLength  bool
	clockShift byte
	widthMode  bool
	divisor    byte
}

func newNoiseGen() *noiseGen {
	return &noiseGen{
		ve: new(volumeEnvelope),
	}
}
func (ng *noiseGen) CurrentSample() float32 {
	if ng.hi {
		return ng.ve.Volume()
	}
	return 0
}

func (ng *noiseGen) Step(frameStep sequencerStep) {
	if (frameStep&ssLength == ssLength) && ng.useLength {
		ng.length--
	}
	if frameStep&ssVolume == ssVolume {
		ng.ve.Step()
	}

	if ng.timerCnt--; ng.timerCnt <= 0 {
		ng.timerCnt = int(divisors[ng.divisor]) << ng.clockShift

		//It has a 15 - bit shift register with feedback.When clocked by the frequency timer, the low two bits(0 and 1) are XORed,
		//all bits are shifted right by one, and the result of the XOR is put into the now - empty high bit.If width mode is 1 (NR43),
		//the XOR result is ALSO put into bit 6 AFTER the shift, resulting in a 7 - bit LFSR.
		//The waveform output is bit 0 of the LFSR, INVERTED.
		result := (ng.lfsr & 0x1) ^ ((ng.lfsr >> 1) & 0x1)
		ng.lfsr >>= 1
		ng.lfsr |= result << 14
		if ng.widthMode {
			ng.lfsr &= 0xBF
			ng.lfsr |= result << 6
		}
		ng.hi = ng.lfsr&0x01 != 0
	}
}

func (ng *noiseGen) Read(addr uint16) byte {
	switch addr {
	case AddrNR41:
		return ng.lengthLoad
	case AddrNR42:
		return ng.ve.Read()
	case AddrNR43:
		var wm byte
		if ng.widthMode {
			wm = 1
		}
		return (ng.clockShift << 4) | (wm << 3) | (ng.divisor & 0x07)
	case AddrNR44:
		var useLen byte
		if ng.useLength {
			useLen = 1
		}
		return (useLen << 6)
	default:
		return 0
	}
}

func (ng *noiseGen) Write(addr uint16, val byte) {
	switch addr {
	case AddrNR41:
		ng.lengthLoad = val & 0x3f
	case AddrNR42:
		ng.ve.Write(val)
	case AddrNR43:
		ng.clockShift = val >> 4
		ng.widthMode = val&0x08 != 0
		ng.divisor = val & 0x07
	case AddrNR44:
		ng.useLength = val&0x40 != 0
		if val&(1<<7) != 0 {
			ng.trigger()
		}
	}
}

func (ng *noiseGen) trigger() {
	ng.length = 64 - ng.lengthLoad
	ng.timerCnt = int(divisors[ng.divisor]) << ng.clockShift
	ng.lfsr = 0x7FFF
	ng.ve.Reset()
}
