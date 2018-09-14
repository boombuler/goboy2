package cpu

func word(bHi, bLo byte) uint16 {
	return uint16(bHi)<<8 | uint16(bLo)
}

func bytes(w uint16) (byte, byte) {
	return byte(w >> 8), byte(w)
}

type pipedOpcode struct {
	codes []opCode
	idx   int
	sub   opCode
}

func (po *pipedOpcode) seek(newIdx int) {
	po.idx = newIdx
	if po.idx < len(po.codes) {
		po.sub = po.codes[po.idx]
	}
}

func (po *pipedOpcode) Exec(cpu *CPU, state *ocState) {
	if po.sub == nil {
		po.seek(0)
	}

	po.sub.Exec(cpu, state)
	po.sub = po.sub.Next(state)
	if po.sub == nil {
		po.seek(po.idx + 1)
	}
}

func (po *pipedOpcode) Next(state *ocState) opCode {
	if po.sub == nil {
		return nil
	}
	return po
}

func (po *pipedOpcode) TakesCycle() bool {
	if po.sub == nil {
		po.seek(0)
	}
	return po.sub.TakesCycle()
}

func pipe(opCodes ...opCode) opCode {
	// flattern piped opcodes:
	newCodes := make([]opCode, 0)
	for _, oc := range opCodes {
		if poc, ok := oc.(*pipedOpcode); ok {
			newCodes = append(newCodes, poc.codes...)
		} else {
			newCodes = append(newCodes, oc)
		}
	}

	return &pipedOpcode{
		codes: newCodes,
	}
}

type opCodeFn func(cpu *CPU, state *ocState)

func (f opCodeFn) Exec(cpu *CPU, state *ocState) {
	f(cpu, state)
}
func (f opCodeFn) Next(state *ocState) opCode {
	return nil
}
func (f opCodeFn) TakesCycle() bool {
	return false
}

type delay struct{}

func (d delay) Exec(cpu *CPU, state *ocState) {}
func (d delay) Next(state *ocState) opCode {
	return nil
}
func (d delay) TakesCycle() bool {
	return true
}

type readByte struct{}

func (ra readByte) Exec(cpu *CPU, state *ocState) {
	addr := state.popW()
	val := cpu.mmu.Read(addr)
	state.pushB(val)
}

func (ra readByte) Next(state *ocState) opCode {
	return nil
}
func (ra readByte) TakesCycle() bool {
	return true
}

type writeByte struct{}

func (ra writeByte) Exec(cpu *CPU, state *ocState) {
	addr := state.popW()
	val := state.popB()

	cpu.mmu.Write(addr, val)
}

func (ra writeByte) Next(state *ocState) opCode {
	return nil
}
func (ra writeByte) TakesCycle() bool {
	return true
}

func writeWord() opCode {
	return pipe(opCodeFn(func(c *CPU, state *ocState) {
		addr := state.popW()
		val := state.popW()
		state.pushB(byte(val))
		state.pushW(addr)

		state.pushB(byte(val >> 8))
		state.pushW(addr + 1)
	}), writeByte{ /*hi*/ }, writeByte{ /*lo*/ })
}

func paramB() opCode {
	return pipe(pc.Read(), opCodeFn(func(c *CPU, s *ocState) {
		adr := s.peekW()
		// Increment PC
		s.pushW(adr + 1)
	}), pc.Write(), readByte{})
}

func paramW() opCode {
	return pipe(paramB(), paramB())
}

type opCodeTable [256]opCode

func (t opCodeTable) Exec(cpu *CPU, state *ocState) {}
func (t opCodeTable) Next(state *ocState) opCode {
	x := state.popB()
	state.pushB(x)
	return t[x]
}

func (t opCodeTable) TakesCycle() bool {
	return false
}
