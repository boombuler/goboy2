package apu

const (
	AddrNR11 uint16 = 0xFF11
	AddrNR12 uint16 = 0xFF12
	AddrNR13 uint16 = 0xFF13
	AddrNR14 uint16 = 0xFF14
)

type soundChannel1 struct {
	*squareWaveGen
}

func newSC1(apu *APU) *soundChannel1 {
	return &soundChannel1{
		newSquareWave(apu, AddrNR11, AddrNR12, AddrNR13, AddrNR14),
	}
}
