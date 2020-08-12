package apu

const (
	waveRAMLen uint16 = 0x0010
)

type waveChannel struct {
	apu        *APU
	dacEnabled bool
	length     int
	timerCnt   int
	sample     float32
	waveRAM    []byte
	pos        int

	useLength  bool
	timerLoad  int
	volume     byte
	lengthLoad byte
	running    bool
}

func waveRAMAddrs() []uint16 {
	res := make([]uint16, 0, 0x10)
	for a := uint16(addrWaveRAM); a < addrWaveRAM+waveRAMLen; a++ {
		res = append(res, a)
	}
	return res
}

func newWaveChannel(apu *APU) *waveChannel {
	return &waveChannel{
		apu:     apu,
		waveRAM: make([]byte, int(waveRAMLen)),
	}
}

func (s *waveChannel) Init(noBoot bool) {
}

func (wc *waveChannel) Reset() {
	wc.dacEnabled = false
	wc.length = 0
	wc.timerCnt = 0
	wc.sample = 0
	wc.pos = 0
	wc.useLength = false
	wc.timerLoad = 0
	wc.volume = 0
	wc.lengthLoad = 0
	wc.running = false
}

func (wc *waveChannel) CurrentSample() float32 {
	return wc.sample
}
func (wc *waveChannel) Step(frameStep sequencerStep) {
	if (frameStep&ssLength == ssLength) && wc.useLength && wc.length > 0 {
		wc.length--
	}

	wc.timerCnt--

	if wc.timerCnt <= 0 {
		wc.reloadTimer()

		if wc.dacEnabled {
			if !wc.useLength || wc.length > 0 {
				wc.pos = (wc.pos + 1) & 0x1F
				idx := wc.pos / 2
				outByte := wc.waveRAM[idx]
				if wc.pos&1 == 0 {
					outByte = outByte >> 4
				}
				outByte = (outByte & 0x0F) >> wc.volumeShift()

				wc.sample = float32(outByte) / 15
			} else {
				wc.running = false
			}
		} else {
			wc.sample = 0
		}
	}
}

func (wc *waveChannel) Active() bool {
	return wc.running
}

func (wc *waveChannel) volumeShift() byte {
	if wc.volume == 0 {
		return 4
	}
	return wc.volume - 1
}

func (wc *waveChannel) reloadTimer() {
	wc.timerCnt = (2048 - wc.timerLoad) / 2
}

func (wc *waveChannel) Read(addr uint16) byte {
	switch addr {
	case addrNR30:
		if wc.dacEnabled {
			return 0xFF
		}
		return 0x7F
	case addrNR31:
		return 0xFF
	case addrNR32:
		return wc.volume<<5 | 0x9F
	case addrNR33:
		return 0xFF
	case addrNR34:
		var lenEnabled byte
		if wc.useLength {
			lenEnabled = 1
		}
		return 0xBF | (lenEnabled << 6)
	default:
		return wc.waveRAM[addr-addrWaveRAM]
	}
}

func (wc *waveChannel) Write(addr uint16, val byte) {
	if !wc.apu.active {
		return
	}

	switch addr {
	case addrNR30:
		wc.dacEnabled = val&0x80 != 0
		if !wc.dacEnabled {
			wc.running = false
		}
	case addrNR31:
		wc.lengthLoad = val
		wc.length = 256 - int(wc.lengthLoad)
	case addrNR32:
		wc.volume = (val >> 5) & 0x03
	case addrNR33:
		wc.timerLoad = (wc.timerLoad & 0x0700) | int(val)
	case addrNR34:
		wc.timerLoad = (wc.timerLoad & 0xFF) | (int(val&0x07) << 8)
		wc.useLength = val&0x40 != 0x00
		if val&0x80 != 0 && wc.dacEnabled {
			wc.trigger()
		}
	default:
		wc.waveRAM[addr-addrWaveRAM] = val
	}
}

func (wc *waveChannel) trigger() {
	wc.running = true
	if wc.length == 0 {
		wc.length = 256
	}
	wc.reloadTimer()
	wc.pos = 0
}
