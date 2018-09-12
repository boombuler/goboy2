package apu

/*
NR30 FF1A E--- ---- DAC power
NR31 FF1B LLLL LLLL Length load (256-L)
NR32 FF1C -VV- ---- Volume code (00=0%, 01=100%, 10=50%, 11=25%)
NR33 FF1D FFFF FFFF Frequency LSB
NR34 FF1E TL-- -FFF Trigger, Length enable, Frequency MSB
*/

const (
	AddrNR30    uint16 = 0xFF1A
	AddrNR31    uint16 = 0xFF1B
	AddrNR32    uint16 = 0xFF1C
	AddrNR33    uint16 = 0xFF1D
	AddrNR34    uint16 = 0xFF1E
	AddrWaveRam uint16 = 0xFF30
	WaveRamLen  uint16 = 0x0010
)

type waveChannel struct {
	active   bool
	length   int
	timerCnt int
	sample   float32
	waveRAM  []byte
	pos      int

	reg34      byte
	reg33      byte
	volume     byte
	lengthLoad byte
}

func waveRamAddrs() []uint16 {
	res := make([]uint16, 0, 0x10)
	for a := uint16(AddrWaveRam); a < AddrWaveRam+WaveRamLen; a++ {
		res = append(res, a)
	}
	return res
}

func newWaveChannel() *waveChannel {
	return &waveChannel{
		waveRAM: make([]byte, int(WaveRamLen)),
	}
}

func (wc *waveChannel) CurrentSample() float32 {
	return wc.sample
}
func (wc *waveChannel) Step(frameStep sequencerStep) {
	useLen := wc.useSoundLength()
	if (frameStep&ssLength == ssLength) && useLen {
		wc.length--
	}

	wc.timerCnt -= 2

	if wc.timerCnt <= 0 {
		wc.timerCnt = 2048 - wc.freq()

		if wc.active {
			if !useLen || wc.length > 0 {
				wc.pos = (wc.pos + 1) & 0x1F
				idx := wc.pos / 2
				outByte := wc.waveRAM[idx]
				if wc.pos&1 == 0 {
					outByte = outByte >> 4
				}
				outByte = (outByte & 0x0F) >> wc.volumeShift()

				wc.sample = float32(outByte) / 15
			}
		} else {
			wc.sample = 0
		}
	}
}

func (wc *waveChannel) volumeShift() byte {
	if wc.volume == 0 {
		return 4
	}
	return wc.volume - 1
}

func (wc *waveChannel) useSoundLength() bool {
	return (wc.reg34 & (1 << 6)) != 0
}

func (wc *waveChannel) freq() int {
	v := int(wc.reg34 & 7)
	v = v<<7 | int(wc.reg33)
	return v
}

func (wc *waveChannel) setLength() {
	wc.length = 256 - int(wc.lengthLoad)
}

func (wc *waveChannel) Read(addr uint16) byte {
	switch addr {
	case AddrNR30:
		if wc.active {
			return 0x80
		}
		return 0x00
	case AddrNR31:
		return wc.lengthLoad
	case AddrNR32:
		return wc.volume << 5
	case AddrNR33:
		return wc.reg33
	case AddrNR34:
		return wc.reg34
	default:
		return wc.waveRAM[addr-AddrWaveRam]
	}
}

func (wc *waveChannel) Write(addr uint16, val byte) {
	switch addr {
	case AddrNR30:
		wc.active = val&0x80 != 0
	case AddrNR31:
		wc.lengthLoad = val
		wc.setLength()
	case AddrNR32:
		wc.volume = (val >> 5) & 0x03
	case AddrNR33:
		wc.reg33 = val
	case AddrNR34:
		wc.reg34 = val
		if val&(1<<7) != 0 {
			wc.setLength()
			wc.pos = 0
		}
	default:
		wc.waveRAM[addr-AddrWaveRam] = val
	}
}
