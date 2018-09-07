package cartridge

import (
	"io"
	"sync"
	"time"
)

type rtcTime struct {
	s byte
	m byte
	h byte
	d uint16
}

type rtc struct {
	t  *rtcTime
	lt *rtcTime

	latch  byte
	halt   bool
	ticker *time.Ticker
	mt     *sync.Mutex
}

func (r *rtc) tick() {
	r.mt.Lock()
	defer r.mt.Unlock()
	if !r.halt {
		if r.t.s++; r.t.s >= 60 {
			r.t.s = 0
			r.t.m++
		}
		if r.t.m >= 60 {
			r.t.m = 0
			r.t.h++
		}
		if r.t.h >= 24 {
			r.t.h = 0
			r.t.d++
		}
	}
}

func newRealTimeClock() *rtc {
	res := new(rtc)
	res.mt = new(sync.Mutex)
	res.t = new(rtcTime)
	res.ticker = time.NewTicker(time.Second)
	res.halt = true
	go func() {
		for _ = range res.ticker.C {
			res.tick()
		}
	}()

	return res
}

func (r *rtc) WriteLatch(val byte) {
	if val == 0x01 && r.latch == 0x00 {
		if r.lt == nil {
			r.mt.Lock()
			defer r.mt.Unlock()
			r.lt = new(rtcTime)
			*r.lt = *r.t
		} else {
			r.lt = nil
		}
	}
	r.latch = val
}

func (r *rtc) Read(bank int) byte {
	r.mt.Lock()
	defer r.mt.Unlock()

	t := *r.t
	if r.lt != nil {
		t = *r.lt
	}

	switch bank {
	case 0x08:
		return t.s
	case 0x09:
		return t.m
	case 0x0A:
		return t.h
	case 0x0B:
		return byte(t.d & 0xFF)
	case 0x0C:
		value := byte(t.d>>8) & 0x01
		if r.halt {
			value |= 0x40
		}
		if t.d>>9 != 0x00 {
			value |= 0x80
		}
		return value
	}
	return 0x00
}

func (r *rtc) Write(bank int, value byte) {
	r.mt.Lock()
	defer r.mt.Unlock()

	switch bank {
	case 0x08:
		r.t.s = value
	case 0x09:
		r.t.m = value
	case 0x0A:
		r.t.h = value
	case 0x0B:
		r.t.d = (r.t.d & 0xFF00) | uint16(value)
	case 0x0C:
		r.halt = value&0x40 == 0x40
		if value&0x80 == 0x00 {
			r.t.d = r.t.d & 0x01FF
		}
		if value&0x01 == 0x01 {
			r.t.d |= 0x0100
		} else {
			r.t.d &= 0xFEFF
		}
	}
}

func (r *rtc) Dump(w io.Writer) {
	r.mt.Lock()
	defer r.mt.Unlock()
	dt := time.Now().Unix()
	w.Write([]byte{
		byte(dt >> 56),
		byte(dt >> 48),
		byte(dt >> 40),
		byte(dt >> 32),
		byte(dt >> 24),
		byte(dt >> 16),
		byte(dt >> 8),
		byte(dt >> 0),
	})

	w.Write([]byte{
		r.t.s, r.t.m, r.t.h,
		byte(r.t.d & 0xFF), byte(r.t.d >> 8),
	})
	if r.halt {
		w.Write([]byte{0x01})
	} else {
		w.Write([]byte{0x00})
	}
}

func (rtc *rtc) Load(r io.Reader) {
	rtc.mt.Lock()
	defer rtc.mt.Unlock()
	dtBuf := make([]byte, 8)
	r.Read(dtBuf)
	dtUnix := int64(0)
	for _, v := range dtBuf {
		dtUnix = dtUnix << 8
		dtUnix = dtUnix | int64(v)
	}
	lastTickDat := time.Unix(dtUnix, 0)
	valBuf := make([]byte, 6)
	r.Read(valBuf)
	rtc.t.s = valBuf[0]
	rtc.t.m = valBuf[1]
	rtc.t.h = valBuf[2]
	rtc.t.d = uint16(valBuf[3]) | (uint16(valBuf[4]) << 8)

	rtc.halt = valBuf[5] == 0x01
	if !rtc.halt {
		dtSec := int64(time.Now().Sub(lastTickDat).Seconds())
		time := dtSec + int64(rtc.t.s)
		rtc.t.s = byte(time % 60)
		time = (time / 60) + int64(rtc.t.m)
		rtc.t.m = byte(time % 60)
		time = (time / 60) + int64(rtc.t.h)
		rtc.t.h = byte(time % 24)
		rtc.t.d = uint16((time / 24) + int64(rtc.t.d))
	}
}
