package apu

import "time"

const (
	sweepInterval = time.Second / 128
)

type volumeEnvelope struct {
	channel soundChannelVolumeEnvelop
	Volume  int
}

type soundChannelVolumeEnvelop interface {
	SoundChannel
	getVolEnvelopCtrl() byte
}

func newVolumeEnvelop(channel soundChannelVolumeEnvelop) *volumeEnvelope {
	return &volumeEnvelope{
		channel: channel,
	}
}

func (ve *volumeEnvelope) increase() bool {
	return (ve.channel.getVolEnvelopCtrl() & (1 << 3)) != 0
}

func (ve *volumeEnvelope) Reset() {
	ve.Volume = int(ve.channel.getVolEnvelopCtrl() >> 4)
}

func (ve *volumeEnvelope) Step() {
	sweepCnt := int(ve.channel.getVolEnvelopCtrl() & 0x07)
	if sweepCnt == 0 {
		return
	}

	if ve.increase() {
		ve.Volume++
	} else {
		ve.Volume--
	}

	if ve.Volume < 0 {
		ve.Volume = 0
	} else if ve.Volume > 15 {
		ve.Volume = 15
	}
}
