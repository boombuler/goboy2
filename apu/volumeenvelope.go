package apu

type volumeEnvelope struct {
	volume     byte
	VolumeLoad byte
	Increase   bool
	period     byte
	periodLoad byte
}

func (ve *volumeEnvelope) dacEnabled() bool {
	return ve.VolumeLoad != 0 || ve.Increase
}

func (ve *volumeEnvelope) reset() {
	ve.volume = 0
	ve.VolumeLoad = 0
	ve.Increase = false
	ve.period = 0
	ve.periodLoad = 0
}

func (ve *volumeEnvelope) Volume() float32 {
	return float32(ve.volume) / 15
}

func (ve *volumeEnvelope) Reset() {
	ve.volume = ve.VolumeLoad
	ve.period = ve.periodLoad
}

func (ve *volumeEnvelope) Step() {
	if ve.periodLoad == 0 {
		return
	}
	if ve.period--; ve.period == 0 {
		ve.period = ve.periodLoad
		if ve.Increase && ve.volume < 15 {
			ve.volume++
		} else if !ve.Increase && ve.volume > 0 {
			ve.volume--
		}
	}
}

func (ve *volumeEnvelope) Write(val byte) {
	ve.VolumeLoad = (val >> 4) & 0x0F
	ve.Increase = (val & 0x08) != 0
	ve.periodLoad = (val & 0x07)
	ve.Reset()
}

func (ve *volumeEnvelope) Read() byte {
	var inc byte
	if ve.Increase {
		inc = 1
	}
	return byte(ve.periodLoad&0x07) | (inc << 3) | (ve.VolumeLoad << 4)
}
