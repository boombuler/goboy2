package apu

/*
void sdlAudioCallback(void*  userdata, void* stream, int len);
*/
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

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
