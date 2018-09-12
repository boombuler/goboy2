package apu

type frameSequencer struct {
	counter int
	curStep byte
}

type sequencerStep byte

const (
	ssOther  sequencerStep = 0
	ssLength sequencerStep = 1 << iota
	ssVolume
	ssSweep
)

func newFrameSequencer() *frameSequencer {
	return &frameSequencer{
		counter: frameSequencerTicks,
		curStep: 0,
	}
}

func (fs *frameSequencer) step() sequencerStep {
	fs.counter--

	stepNo := byte(0xFF)
	if fs.counter == 0 {
		stepNo = fs.curStep
		fs.curStep = (stepNo + 1) % 8
		fs.counter = frameSequencerTicks
	}

	result := ssOther
	if stepNo%2 == 0 {
		result |= ssLength
	}
	if stepNo == 2 || stepNo == 6 {
		result |= ssSweep
	}
	if stepNo == 7 {
		result |= ssVolume
	}
	return result
}
