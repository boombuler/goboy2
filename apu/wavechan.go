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
	length   byte
	timerCnt int
	sample   float32
	waveRam  []byte
	pos      int

	reg34 byte
	reg33 byte
	reg32 byte
	reg31 byte
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
		waveRam: make([]byte, int(WaveRamLen)),
	}
}

func (wc *waveChannel) CurrentSample() float32 {
	return wc.sample
}
func (wc *waveChannel) Step(frameStep byte) {
	useLen := wc.useSoundLength()
	if frameStep%2 == 0 && useLen {
		wc.length--
	}

	wc.timerCnt -= 2

	if wc.timerCnt <= 0 {
		wc.timerCnt = 2048 - wc.freq()

		if wc.active {
			if !useLen || wc.length > 0 {
				wc.pos = (wc.pos + 1) & 0x1F
				idx := wc.pos / 2
				outByte := wc.waveRam[idx]
				if wc.pos&1 == 0 {
					outByte = outByte >> 4
				}
				outByte &= 0xF

				wc.sample = (float32(outByte) / 15) * wc.volume()
			}
		} else {
			wc.sample = 0
		}
	}
}

func (wc *waveChannel) volume() float32 {
	switch (wc.reg32 >> 4) & 0x03 {
	case 1:
		return 1
	case 2:
		return 0.5
	case 3:
		return 0.25
	default:
		return 0
	}
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
	wc.length = 64 - (wc.reg31 & 0x3F)
}

func (wc *waveChannel) Read(addr uint16) byte {
	switch addr {
	case AddrNR30:
		if wc.active {
			return 0x80
		}
		return 0x00
	case AddrNR31:
		return wc.reg31
	case AddrNR32:
		return wc.reg32
	case AddrNR33:
		return wc.reg33
	case AddrNR34:
		return wc.reg34
	default:
		return wc.waveRam[addr-AddrWaveRam]
	}
}

func (wc *waveChannel) Write(addr uint16, val byte) {
	switch addr {
	case AddrNR30:
		wc.active = val&0x80 != 0
	case AddrNR31:
		wc.reg31 = val
		wc.setLength()
	case AddrNR32:
		wc.reg32 = val
	case AddrNR33:
		wc.reg33 = val
	case AddrNR34:
		wc.reg34 = val
		if val&(1<<7) != 0 {
			wc.setLength()
			wc.pos = 0
		}
	default:
		wc.waveRam[addr-AddrWaveRam] = val
	}
}
