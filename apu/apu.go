package apu

/*
void sdlAudioCallback(void*  userdata, void* stream, int len);
*/
import "C"

import (
	"fmt"
	"goboy2/mmu"
	"reflect"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	gbTicksPerSecond               = 4194304 / 4
	sampleRate                     = 44100
	sampleBufferSize               = 2048
	sampleDuration   time.Duration = time.Second / sampleRate
	stepDuration     time.Duration = time.Second / gbTicksPerSecond
)

type APU struct {
	volume      float64
	soundBuffer []float32

	bufIdx  int
	sampleT time.Duration

	generators []SoundChannel
	format     sdl.AudioFormat
}

func New(mmu mmu.MMU) *APU {
	apu := &APU{
		volume:      0,
		soundBuffer: make([]float32, sampleBufferSize, sampleBufferSize),
	}
	ch2 := &soundChannel2{apu: apu}
	apu.generators = []SoundChannel{
		ch2,
	}
	mmu.AddIODevice(ch2, AddrNR21, AddrNR22, AddrNR23, AddrNR24)
	return apu
}

type SoundChannel interface {
	CurrentSample() float32
	Step()
}

var (
	currentAPU *APU
)

//export sdlAudioCallback
func sdlAudioCallback(a unsafe.Pointer, stream unsafe.Pointer, l C.int) {
	apu := currentAPU
	length := int(l) / 2
	outStream := *(*[]float32)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(stream),
		Len:  length,
		Cap:  length,
	}))
	idx := apu.bufIdx
	apu.bufIdx = 0
	for i := idx; i < length; i++ {
		apu.soundBuffer[i] = 0 // clear unfilled buffer
	}

	copy(outStream, apu.soundBuffer)
}

func (apu *APU) Step() {
	var sum float32

	apu.sampleT += stepDuration
	sampleStep := apu.sampleT >= sampleDuration
	for _, sc := range apu.generators {
		sc.Step()
		if sampleStep {
			sum = sum + sc.CurrentSample()
		}
	}

	if sampleStep {
		apu.sampleT -= sampleDuration
		apu.bufIdx++
		apu.soundBuffer[apu.bufIdx] = sum / float32(len(apu.generators))
	}
}

// Start audio playback
func (apu *APU) Start() error {
	if err := sdl.InitSubSystem(sdl.INIT_AUDIO); err != nil {
		return err
	}

	var wanted sdl.AudioSpec
	wanted.Freq = sampleRate
	wanted.Format = sdl.AUDIO_F32SYS
	wanted.Channels = 1
	wanted.Samples = sampleBufferSize
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
