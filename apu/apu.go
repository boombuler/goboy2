package apu

import (
	"goboy2/consts"
	"goboy2/mmu"
	"sync"
	"time"
)

type audioChannel bool

const (
	left  audioChannel = true
	right audioChannel = false
)

const (
	frameSequencerTicks               = consts.TicksPerSecond / 512
	sampleRate                        = 2 * 22050
	sampleBufferLength                = 1024
	channelCount                      = 2
	sampleDuration      time.Duration = time.Second / sampleRate
	stepDuration        time.Duration = time.Second / consts.TicksPerSecond
	sampleSize                        = 4 // sizeOf(float32)

	addrNR10    uint16 = 0xFF10
	addrNR11    uint16 = 0xFF11
	addrNR12    uint16 = 0xFF12
	addrNR13    uint16 = 0xFF13
	addrNR14    uint16 = 0xFF14
	addrNR21    uint16 = 0xFF16
	addrNR22    uint16 = 0xFF17
	addrNR23    uint16 = 0xFF18
	addrNR24    uint16 = 0xFF19
	addrNR30    uint16 = 0xFF1A
	addrNR31    uint16 = 0xFF1B
	addrNR32    uint16 = 0xFF1C
	addrNR33    uint16 = 0xFF1D
	addrNR34    uint16 = 0xFF1E
	addrNR41    uint16 = 0xFF20
	addrNR42    uint16 = 0xFF21
	addrNR43    uint16 = 0xFF22
	addrNR44    uint16 = 0xFF23
	addrNR50    uint16 = 0xFF24
	addrNR51    uint16 = 0xFF25
	addrNR52    uint16 = 0xFF26
	addrWaveRAM uint16 = 0xFF30
)

// APU implements a gameboy audio processing unit
type APU struct {
	mmu          mmu.MMU
	TestMode     bool
	masterVolume float32
	soundBuffer  []float32
	m            *sync.Mutex
	fs           *frameSequencer

	sampleT     time.Duration
	sampleLeft  float32
	sampleRight float32

	volumeSelect  byte
	channelSelect byte
	active        bool

	generators []soundChannel
}

// New creates a new gameboy APU
func New(mmu mmu.MMU) *APU {
	apu := &APU{
		masterVolume: 0.3,
		mmu:          mmu,
		m:            new(sync.Mutex),
		soundBuffer:  make([]float32, 0),
		fs:           newFrameSequencer(),
	}
	ch1 := newSweepSquareWaveGen(apu)
	ch2 := newSquareWave(apu, addrNR21, addrNR22, addrNR23, addrNR24)
	ch3 := newWaveChannel()
	ch4 := newNoiseGen()

	apu.generators = []soundChannel{
		ch1,
		ch2,
		ch3,
		ch4,
	}
	mmu.AddIODevice(apu, addrNR50, addrNR51, addrNR52)
	mmu.AddIODevice(ch1, addrNR10, addrNR11, addrNR12, addrNR13, addrNR14)
	mmu.AddIODevice(ch2, addrNR21, addrNR22, addrNR23, addrNR24)
	mmu.AddIODevice(ch3, addrNR30, addrNR31, addrNR32, addrNR33, addrNR34)
	mmu.AddIODevice(ch3, waveRAMAddrs()...)
	mmu.AddIODevice(ch4, addrNR41, addrNR42, addrNR43, addrNR44)
	apu.reset()
	return apu
}

type soundChannel interface {
	CurrentSample() float32
	Step(s sequencerStep)
	Reset()
	Active() bool
}

func (apu *APU) reset() {
	apu.volumeSelect = 0
	apu.channelSelect = 0
	apu.sampleLeft = 0
	apu.sampleRight = 0
	apu.active = false
	apu.fs.reset()
	for _, ch := range apu.generators {
		ch.Reset()
	}
}

func (apu *APU) Read(addr uint16) byte {
	switch addr {
	case addrNR50:
		return apu.volumeSelect
	case addrNR51:
		return apu.channelSelect
	case addrNR52:
		result := byte(0x70)
		if apu.active {
			result |= 0x80
		}
		for i, ch := range apu.generators {
			if ch.Active() {
				result |= (1 << byte(i))
			}
		}
		return result
	default:
		return 0
	}
}

func (apu *APU) Write(addr uint16, val byte) {
	switch addr {
	case addrNR50:
		apu.volumeSelect = val
	case addrNR51:
		apu.channelSelect = val
	case addrNR52:
		oldActive := apu.active
		apu.active = val&0x80 != 0
		if oldActive && !apu.active {
			apu.reset()
		}
	}
}

func mix(a, b float32) float32 {
	return a + b - (a * b)
}

// Step Executes the next apu step
func (apu *APU) Step() {
	var sampleLeft float32
	var sampleRight float32
	apu.sampleT += stepDuration

	if !apu.active {
		apu.sampleLeft = 0
		apu.sampleRight = 0

		if apu.sampleT >= sampleDuration {
			apu.sampleT -= sampleDuration
			apu.m.Lock()
			apu.soundBuffer = append(apu.soundBuffer, sampleLeft, sampleRight)
			sampleCount := len(apu.soundBuffer)
			apu.m.Unlock()
			if sampleCount > sampleBufferLength*channelCount*2 {
				sleepTime := sampleDuration * sampleBufferLength
				time.Sleep(sleepTime)
			}
		}
		return
	}

	step := apu.fs.step()

	for i, sc := range apu.generators {
		sc.Step(step)
		sample := sc.CurrentSample()
		sampleLeft += (sample * apu.getVolume(left, i))
		sampleRight += (sample * apu.getVolume(right, i))
	}

	genCount := float32(len(apu.generators))
	apu.sampleLeft = mix(apu.sampleLeft, (sampleLeft / genCount))
	apu.sampleRight = mix(apu.sampleRight, (sampleLeft / genCount))

	if apu.sampleT >= sampleDuration {
		apu.sampleT -= sampleDuration

		sampleLeft = apu.sampleLeft * apu.masterVolume
		sampleRight = apu.sampleRight * apu.masterVolume
		apu.sampleLeft = 0
		apu.sampleRight = 0

		apu.m.Lock()
		apu.soundBuffer = append(apu.soundBuffer, sampleLeft, sampleRight)
		sampleCount := len(apu.soundBuffer)
		apu.m.Unlock()

		if (sampleCount > sampleBufferLength*channelCount*2) && !apu.TestMode {
			sleepTime := sampleDuration * sampleBufferLength
			time.Sleep(sleepTime)
		}
	}
}

func (apu *APU) getVolume(ch audioChannel, sc int) float32 {
	if !apu.active || apu.TestMode {
		return 0
	}
	var soShift uint
	if ch == left {
		soShift += 4
	}
	if enabled := (apu.channelSelect>>(soShift+uint(sc)))&1 != 0; !enabled {
		return 0
	}
	return float32((apu.volumeSelect>>soShift)&0x3) / 7
}
