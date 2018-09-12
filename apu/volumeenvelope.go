package apu

import "time"

const (
	sweepInterval = time.Second / 128
)

type volumeEnvelope struct {
	volume     byte
	VolumeLoad byte
	Increase   bool
	period     byte
	periodLoad byte
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
