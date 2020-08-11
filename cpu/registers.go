package cpu

type flag byte

func (f flag) String() string {
	res := ""
	if f&zero == zero {
		res += "Z"
	} else {
		res += "-"
	}
	if f&substract == substract {
		res += "N"
	} else {
		res += "-"
	}
	if f&halfcarry == halfcarry {
		res += "H"
	} else {
		res += "-"
	}
	if f&carry == carry {
		res += "C"
	} else {
		res += "-"
	}
	return res
}

const (
	zero      flag = 1 << 7 // Z
	substract flag = 1 << 6 // N
	halfcarry flag = 1 << 5 // H
	carry     flag = 1 << 4 // C
)

type registers struct {
	a, b, c, d, e, h, l byte
	pc, sp              uint16
	f                   flag
}

type reader interface {
	Read() opCode
}

type writer interface {
	Write() opCode
}

type readerwriter interface {
	reader
	writer
}

type reg8 byte

const (
	a reg8 = iota
	b
	c
	d
	e
	f
	h
	l
)

func (r reg8) Read() opCode {
	switch r {
	case a:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushB(cpu.a)
		})
	case b:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushB(cpu.b)
		})
	case c:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushB(cpu.c)
		})
	case d:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushB(cpu.d)
		})
	case e:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushB(cpu.e)
		})
	case f:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushB(byte(cpu.f))
		})
	case h:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushB(cpu.h)
		})
	case l:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushB(cpu.l)
		})
	default:
		panic("Invalid Register.")
	}
}

func (r reg8) Write() opCode {
	switch r {
	case a:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			cpu.a = state.popB()
		})
	case b:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			cpu.b = state.popB()
		})
	case c:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			cpu.c = state.popB()
		})
	case d:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			cpu.d = state.popB()
		})
	case e:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			cpu.e = state.popB()
		})
	case f:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			val := state.popB()
			cpu.f = flag(val) & (zero | carry | halfcarry | substract)
		})
	case h:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			cpu.h = state.popB()
		})
	case l:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			cpu.l = state.popB()
		})
	default:
		panic("Invalid Register.")
	}
}

type reg16 byte

const (
	pc reg16 = iota
	sp
	af
	bc
	de
	hl
)

func (r reg16) Read() opCode {
	switch r {
	case pc:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushW(cpu.pc)
		})
	case sp:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushW(cpu.sp)
		})
	case af:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushW(word(cpu.a, byte(cpu.f)))
		})
	case bc:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushW(word(cpu.b, cpu.c))
		})
	case de:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushW(word(cpu.d, cpu.e))
		})
	case hl:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			state.pushW(word(cpu.h, cpu.l))
		})
	default:
		panic("Invalid Register.")
	}
}

func (r reg16) Write() opCode {
	switch r {
	case pc:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			newPC := state.popW()
			if cpu.haltBug {
				cpu.haltBug = false
			} else {
				cpu.pc = newPC
			}
		})
	case sp:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			cpu.sp = state.popW()
		})
	case af:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			wrd := state.popW()
			a, f := bytes(wrd)
			cpu.a = a
			cpu.f = flag(f) & (zero | carry | halfcarry | substract)
		})
	case bc:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			wrd := state.popW()
			cpu.b, cpu.c = bytes(wrd)
		})
	case de:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			wrd := state.popW()
			cpu.d, cpu.e = bytes(wrd)
		})
	case hl:
		return opCodeFn(func(cpu *CPU, state *ocState) {
			wrd := state.popW()
			cpu.h, cpu.l = bytes(wrd)
		})
	default:
		panic("Invalid Register.")
	}
}

type reg16Ref reg16

func (r reg16Ref) Write() opCode {
	return pipe(reg16(r).Read(), writeByte{})
}

func (r reg16Ref) Read() opCode {
	return pipe(reg16(r).Read(), readByte{})
}

func (r reg16) Deref() readerwriter {
	return reg16Ref(r)
}
