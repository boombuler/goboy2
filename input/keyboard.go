package input

import (
	"sync"

	"github.com/boombuler/goboy2/consts"
	"github.com/boombuler/goboy2/mmu"

	"github.com/veandco/go-sdl2/sdl"
)

type KeyMap struct {
	Up     sdl.Keycode
	Left   sdl.Keycode
	Down   sdl.Keycode
	Right  sdl.Keycode
	A      sdl.Keycode
	B      sdl.Keycode
	Start  sdl.Keycode
	Select sdl.Keycode
}

var DefaultKeymap = KeyMap{
	Up:     sdl.K_UP,
	Left:   sdl.K_LEFT,
	Right:  sdl.K_RIGHT,
	Down:   sdl.K_DOWN,
	Start:  sdl.K_RETURN,
	Select: sdl.K_BACKSPACE,
	A:      sdl.Keycode('x'),
	B:      sdl.Keycode('y'),
}

type Keyboard struct {
	mmu       mmu.MMU
	lock      *sync.Mutex
	keyState  [2]byte
	keyMap    KeyMap
	colSelect byte
}

func NewKeyboard(m mmu.MMU) *Keyboard {
	kb := new(Keyboard)
	kb.mmu = m
	kb.lock = new(sync.Mutex)
	kb.keyMap = DefaultKeymap
	kb.keyState[0], kb.keyState[1] = 0x0F, 0x0F
	m.AddIODevice(kb, consts.AddrInput)
	return kb
}

const (
	col1 byte = 0x10
	col2 byte = 0x20
)

func (kb *Keyboard) Read(addr uint16) byte {
	var fixedMask byte = 0xC0
	if kb.mmu.HardwareCompat() == consts.GBC {
		fixedMask = 0xF0
	}

	if addr == consts.AddrInput {
		kb.lock.Lock()
		defer kb.lock.Unlock()

		switch kb.colSelect {
		case col1:
			return kb.keyState[1] | fixedMask
		case col2:
			return kb.keyState[0] | fixedMask
		default:
			return kb.keyState[0] | kb.keyState[1] | fixedMask
		}
	}
	return 0x00
}
func (kb *Keyboard) Write(addr uint16, value byte) {
	if addr == consts.AddrInput {
		kb.colSelect = value & 0x30
	}
}

func (kb *Keyboard) HandleKeyEvent(isPressed bool, key sdl.Keycode) {
	kb.lock.Lock()
	defer kb.lock.Unlock()

	if isPressed {
		switch key {
		case kb.keyMap.Right:
			kb.keyState[0] &^= 0x1
		case kb.keyMap.Left:
			kb.keyState[0] &^= 0x2
		case kb.keyMap.Up:
			kb.keyState[0] &^= 0x4
		case kb.keyMap.Down:
			kb.keyState[0] &^= 0x8
		case kb.keyMap.A:
			kb.keyState[1] &^= 0x1
		case kb.keyMap.B:
			kb.keyState[1] &^= 0x2
		case kb.keyMap.Select:
			kb.keyState[1] &^= 0x4
		case kb.keyMap.Start:
			kb.keyState[1] &^= 0x8
		}
		kb.mmu.RequestInterrupt(mmu.IRQJoypad)
	} else {
		switch key {
		case kb.keyMap.Right:
			kb.keyState[0] |= 0x1
		case kb.keyMap.Left:
			kb.keyState[0] |= 0x2
		case kb.keyMap.Up:
			kb.keyState[0] |= 0x4
		case kb.keyMap.Down:
			kb.keyState[0] |= 0x8
		case kb.keyMap.A:
			kb.keyState[1] |= 0x1
		case kb.keyMap.B:
			kb.keyState[1] |= 0x2
		case kb.keyMap.Select:
			kb.keyState[1] |= 0x4
		case kb.keyMap.Start:
			kb.keyState[1] |= 0x8
		}
	}
}
