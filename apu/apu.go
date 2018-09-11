package apu

/*
void sdlAudioCallback(void*  userdata, void* stream, int len);
*/
import "C"

import (
	"fmt"
	"goboy2/mmu"
	"reflect"
	"sync"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type audioChannel bool

const (
	left  audioChannel = true
	right audioChannel = false
)

const (
	gbFreq                            = 4194304
	gbTicksPerSecond                  = gbFreq / 4
	frameSequencerTicks               = gbTicksPerSecond / 512
	sampleRate                        = 2 * 22050
	sampleBufferLength                = 4 * 1024
	channelCount                      = 2
	sampleDuration      time.Duration = time.Second / sampleRate
	stepDuration        time.Duration = time.Second / gbTicksPerSecond
	sampleSize                        = 4 // sizeOf(float32)

	AddrNR21 uint16 = 0xFF16
	AddrNR22 uint16 = 0xFF17
	AddrNR23 uint16 = 0xFF18
	AddrNR24 uint16 = 0xFF19

	AddrNR50 uint16 = 0xFF24
	AddrNR51 uint16 = 0xFF25
	AddrNR52 uint16 = 0xFF26
)

type APU struct {
	masterVolume float32
	soundBuffer  []float32
	m            *sync.Mutex

	frameSeq     int
	frameSeqStep byte
	sampleT      time.Duration
	sampleCnt    int
	sampleLeft   float32
	sampleRight  float32

	volumeSelect  byte
	channelSelect byte
	active        bool

	generators []SoundChannel
	format     sdl.AudioFormat
}

func New(mmu mmu.MMU) *APU {
	apu := &APU{
		masterVolume: 0.3,
		m:            new(sync.Mutex),
		soundBuffer:  make([]float32, 0),
		frameSeq:     frameSequencerTicks,
	}
	ch1 := newSC1(apu)
	ch2 := newSquareWave(apu, AddrNR21, AddrNR22, AddrNR23, AddrNR24)

	apu.generators = []SoundChannel{
		ch1,
		ch2,
	}
	mmu.AddIODevice(apu, AddrNR50, AddrNR51, AddrNR52)
	mmu.AddIODevice(ch1, AddrNR10, AddrNR11, AddrNR12, AddrNR13, AddrNR14)
	mmu.AddIODevice(ch2, AddrNR21, AddrNR22, AddrNR23, AddrNR24)
	return apu
}

type SoundChannel interface {
	CurrentSample() float32
	Step(s byte)
}

var (
	currentAPU *APU
)

//export sdlAudioCallback
func sdlAudioCallback(a unsafe.Pointer, stream unsafe.Pointer, l C.int) {
	apu := currentAPU
	length := int(l) / sampleSize
	outStream := *(*[]float32)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(stream),
		Len:  length,
		Cap:  length,
	}))
	apu.m.Lock()
	bufLen := len(apu.soundBuffer)
	copy(outStream, apu.soundBuffer)

	if bufLen > length {
		bufSize := bufLen - length
		for i := 0; i < bufSize; i++ {
			apu.soundBuffer[i] = apu.soundBuffer[i+length]
		}
		apu.soundBuffer = apu.soundBuffer[:bufSize]
	} else {
		if bufLen < length {
			for i := bufLen; i < length; i++ {
				outStream[i] = 0 // unfilled buffer
			}
		}
		apu.soundBuffer = apu.soundBuffer[:0]
	}
	apu.m.Unlock()
}

func (apu *APU) Read(addr uint16) byte {
	switch addr {
	case AddrNR50:
		return apu.volumeSelect
	case AddrNR51:
		return apu.channelSelect
	case AddrNR52:
		if apu.active {
			return 0x8F // todo...
		}
		return 0x00
	default:
		return 0
	}
}

func (apu *APU) Write(addr uint16, val byte) {
	switch addr {
	case AddrNR50:
		apu.volumeSelect = val
	case AddrNR51:
		apu.channelSelect = val
	case AddrNR52:
		apu.active = val&0x80 != 0
	}
}

func (apu *APU) Step() {
	var sampleLeft float32
	var sampleRight float32
	apu.sampleT += stepDuration

	apu.frameSeq--
	var step byte = 0xFF
	if apu.frameSeq == 0 {
		step = apu.frameSeqStep
	}

	for i, sc := range apu.generators {
		sc.Step(step)
		sample := sc.CurrentSample()
		sampleLeft += (sample * apu.getVolume(left, i))
		sampleRight += (sample * apu.getVolume(right, i))
	}

	if step != 0xFF {
		apu.frameSeqStep = (apu.frameSeqStep + 1) % 8
		apu.frameSeq = frameSequencerTicks
	}

	apu.sampleCnt++
	apu.sampleLeft += (sampleLeft / float32(len(apu.generators))) * apu.masterVolume
	apu.sampleRight += (sampleRight / float32(len(apu.generators))) * apu.masterVolume

	if apu.sampleT >= sampleDuration {
		apu.sampleT -= sampleDuration

		sampleLeft = apu.sampleLeft / float32(apu.sampleCnt)
		sampleRight = apu.sampleRight / float32(apu.sampleCnt)
		apu.sampleCnt = 0
		apu.sampleLeft = 0
		apu.sampleRight = 0

		apu.m.Lock()
		apu.soundBuffer = append(apu.soundBuffer, sampleLeft, sampleRight)
		sampleCount := len(apu.soundBuffer)
		apu.m.Unlock()

		if sampleCount > sampleBufferLength*channelCount*2 {
			sleepTime := sampleDuration * sampleBufferLength
			time.Sleep(sleepTime)
		}
	}
}

func (apu *APU) getVolume(ch audioChannel, sc int) float32 {
	if !apu.active {
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

// Start audio playback
func (apu *APU) Start() error {
	if err := sdl.InitSubSystem(sdl.INIT_AUDIO); err != nil {
		return err
	}

	var wanted sdl.AudioSpec
	wanted.Freq = sampleRate
	wanted.Format = sdl.AUDIO_F32SYS
	wanted.Channels = channelCount
	wanted.Samples = sampleBufferLength
	wanted.Callback = (sdl.AudioCallback)(unsafe.Pointer(C.sdlAudioCallback))

	var have sdl.AudioSpec
	if err := sdl.OpenAudio(&wanted, &have); err != nil {
		return err
	} else if wanted.Format != have.Format {
		sdl.CloseAudio()
		return fmt.Errorf("unsupported audio format: %v", have.Format)
	}
	apu.format = have.Format
	currentAPU = apu
	sdl.PauseAudio(false) // start audio playing.
	return nil
}

// Stop audio playback
func (apu *APU) Stop() {
	sdl.CloseAudio()
	if currentAPU == apu {
		currentAPU = nil
	}
}
