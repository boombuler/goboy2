package apu

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

	ng.timerCnt--

	if ng.timerCnt <= 0 {
		ng.timerCnt = int(divisors[ng.divisor]) << ng.clockShift

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
	case addrNR41:
		return ng.lengthLoad
	case addrNR42:
		return ng.ve.Read()
	case addrNR43:
		var wm byte
		if ng.widthMode {
			wm = 1
		}
		return (ng.clockShift << 4) | (wm << 3) | (ng.divisor & 0x07)
	case addrNR44:
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
	case addrNR41:
		ng.lengthLoad = val & 0x3f
	case addrNR42:
		ng.ve.Write(val)
	case addrNR43:
		ng.clockShift = val >> 4
		ng.widthMode = val&0x08 != 0
		ng.divisor = val & 0x07
	case addrNR44:
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
