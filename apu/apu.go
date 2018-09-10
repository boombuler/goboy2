package apu

/*
void sdlAudioCallback(void*  userdata, void* stream, int len);
*/
import "C"

import (
	"fmt"
	"goboy2/mmu"
	"reflect"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	freq       = 44100
	samplerate = 2048
)

type APU struct {
	volume      float64
	soundBuffer []int16

	generators []SoundChannel
	format     sdl.AudioFormat
}

func New(mmu mmu.MMU) *APU {
	apu := &APU{
		volume:      0,
		soundBuffer: make([]int16, samplerate, samplerate),
	}
	ch2 := &soundChannel2{apu: apu}
	apu.generators = []SoundChannel{
		ch2,
	}
	mmu.AddIODevice(ch2, AddrNR21, AddrNR22, AddrNR23, AddrNR24)
	return apu
}

type SoundChannel interface {
	GenerateSamples(buffer []int16) bool
	Step()
}

var (
	currentAPU *APU
)

//export sdlAudioCallback
func sdlAudioCallback(a unsafe.Pointer, stream unsafe.Pointer, l C.int) {
	apu := currentAPU
	length := int(l) / 2
	outStream := *(*[]int16)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(stream),
		Len:  length,
		Cap:  length,
	}))

	for i := int(0); i < length; i++ {
		outStream[i] = 0
	}

	//volume := int(apu.volume * float64(sdl.MIX_MAXVOLUME))

	activeChannelCnt := 0
	sampleBuffer := outStream
	for _, gen := range apu.generators {
		if gen.GenerateSamples(sampleBuffer) {
			if activeChannelCnt == 1 {
				sampleBuffer = apu.soundBuffer
			}
			activeChannelCnt++
		}
		if activeChannelCnt > 1 {
			src := (*uint8)(unsafe.Pointer(&sampleBuffer[0]))
			sdl.MixAudioFormat((*uint8)(stream), src, apu.format, uint32(l), sdl.MIX_MAXVOLUME)
		}
	}
}

func (apu *APU) Step() {
	for _, sc := range apu.generators {
		sc.Step()
	}
}

// Start audio playback
func (apu *APU) Start() error {
	if err := sdl.InitSubSystem(sdl.INIT_AUDIO); err != nil {
		return err
	}

	var wanted sdl.AudioSpec
	wanted.Freq = freq
	wanted.Format = sdl.AUDIO_S16SYS
	wanted.Channels = 1
	wanted.Samples = samplerate
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
