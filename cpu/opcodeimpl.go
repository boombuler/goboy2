package cpu

import (
	"github.com/boombuler/goboy2/consts"
	"github.com/boombuler/goboy2/mmu"
)

func stop() opCode {
	return pipe(opCodeFn(func(c *CPU, s *ocState) {
		if c.key1 != nil {
			if c.key1.changeSpeed() {
				s.pushB(1)
				return
			}
		}
		s.pushB(0)
	}), opCodeFn(func(c *CPU, s *ocState) {
		stopAlreadyHandled := s.popB() != 0
		if !stopAlreadyHandled {
			// TODO: implement original STOP
		}
	}))
}

func halt() opCode {
	return opCodeFn(func(c *CPU, s *ocState) {
		if c.ime || (mmu.IRQ(c.mmu.Read(consts.AddrIRQFlags))&mmu.IRQ(c.mmu.Read(consts.AddrIRQEnabled))&mmu.IRQAll) == mmu.IRQNone {
			c.haltEnabled = true
		} else {
			c.haltBug = true
		}
	})
}

func invalid() opCode {
	return nop()
}

func nop() opCode {
	return opCodeFn(func(c *CPU, s *ocState) {})
}

func ld(out opCode, in opCode) opCode {
	return pipe(in, out)
}

func incR16(r reg16) opCode {
	return pipe(
		delay{},
		r.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushW(s.popW() + 1)
		}),
		r.Write(),
	)
}

func decR16(r reg16) opCode {
	return pipe(
		delay{},
		r.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushW(s.popW() - 1)
		}),
		r.Write(),
	)
}

func incR8(r readerwriter) opCode {
	return pipe(
		r.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushB(c.incByte(s.popB()))
		}),
		r.Write(),
	)
}

func decR8(r readerwriter) opCode {
	return pipe(
		r.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushB(c.decByte(s.popB()))
		}),
		r.Write(),
	)
}

func rlca() opCode {
	return pipe(
		a.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushB(c.rlc(s.popB(), false))
		}),
		a.Write(),
	)
}

func rlc(rw readerwriter) opCode {
	return pipe(
		rw.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushB(c.rlc(s.popB(), true))
		}),
		rw.Write(),
	)
}

func rrc(rw readerwriter) opCode {
	return pipe(
		rw.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushB(c.rrc(s.popB(), true))
		}),
		rw.Write(),
	)
}

func rl(rw readerwriter) opCode {
	return pipe(
		rw.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushB(c.rl(s.popB(), true))
		}),
		rw.Write(),
	)
}

func rr(rw readerwriter) opCode {
	return pipe(
		rw.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushB(c.rr(s.popB(), true))
		}),
		rw.Write(),
	)
}

func rla() opCode {
	return pipe(
		a.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushB(c.rl(s.popB(), false))
		}),
		a.Write(),
	)
}

func rrca() opCode {
	return pipe(
		a.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushB(c.rrc(s.popB(), false))
		}),
		a.Write(),
	)
}

func rra() opCode {
	return pipe(
		a.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			s.pushB(c.rr(s.popB(), false))
		}),
		a.Write(),
	)
}

func addR16(r1 reg16, r2 reg16) opCode {
	return pipe(
		delay{},
		r1.Read(),
		r2.Read(),
		opCodeFn(func(c *CPU, state *ocState) {
			v2 := state.popW()
			v1 := state.popW()
			state.pushW(c.addWords(v1, v2))
		}),
		r1.Write(),
	)
}

func addR8(r1 readerwriter, r2 opCode) opCode {
	return pipe(
		r1.Read(),
		r2,
		opCodeFn(func(c *CPU, state *ocState) {
			v2 := state.popB()
			v1 := state.popB()
			state.pushB(c.addBytes(v1, v2))
		}),
		r1.Write(),
	)
}

func adcR8(r1 readerwriter, r2 opCode) opCode {
	return pipe(
		r1.Read(),
		r2,
		opCodeFn(func(c *CPU, state *ocState) {
			v2 := state.popB()
			v1 := state.popB()
			state.pushB(c.adcBytes(v1, v2))
		}),
		r1.Write(),
	)
}

func subR8(r1 readerwriter, r2 opCode) opCode {
	return pipe(
		r1.Read(),
		r2,
		opCodeFn(func(c *CPU, state *ocState) {
			v2 := state.popB()
			v1 := state.popB()
			state.pushB(c.subBytes(v1, v2))
		}),
		r1.Write(),
	)
}

func sbcR8(r1 readerwriter, r2 opCode) opCode {
	return pipe(
		r1.Read(),
		r2,
		opCodeFn(func(c *CPU, state *ocState) {
			v2 := state.popB()
			v1 := state.popB()
			state.pushB(c.sbcBytes(v1, v2))
		}),
		r1.Write(),
	)
}

func andR8(r1 readerwriter, r2 opCode) opCode {
	return pipe(
		r1.Read(),
		r2,
		opCodeFn(func(c *CPU, state *ocState) {
			v2 := state.popB()
			v1 := state.popB()
			state.pushB(c.andBytes(v1, v2))
		}),
		r1.Write(),
	)
}

func orR8(r1 readerwriter, r2 opCode) opCode {
	return pipe(
		r1.Read(),
		r2,
		opCodeFn(func(c *CPU, state *ocState) {
			v2 := state.popB()
			v1 := state.popB()
			state.pushB(c.orBytes(v1, v2))
		}),
		r1.Write(),
	)
}

func xorR8(r1 readerwriter, r2 opCode) opCode {
	return pipe(
		r1.Read(),
		r2,
		opCodeFn(func(c *CPU, state *ocState) {
			v2 := state.popB()
			v1 := state.popB()
			state.pushB(c.xorBytes(v1, v2))
		}),
		r1.Write(),
	)
}

func cpR8(r1 readerwriter, r2 opCode) opCode {
	return pipe(
		r1.Read(),
		r2,
		opCodeFn(func(c *CPU, state *ocState) {
			v2 := state.popB()
			v1 := state.popB()
			c.subBytes(v1, v2)
		}),
	)
}

func incHLFast() opCode {
	return opCodeFn(func(c *CPU, s *ocState) {
		c.h, c.l = bytes(word(c.h, c.l) + 1)
	})
}

func decHLFast() opCode {
	return opCodeFn(func(c *CPU, s *ocState) {
		c.h, c.l = bytes(word(c.h, c.l) - 1)
	})
}

func daa() opCode {
	return opCodeFn(func(c *CPU, s *ocState) {
		c.daa()
	})
}

func cpl() opCode {
	return opCodeFn(func(c *CPU, s *ocState) {
		c.a = ^c.a
		c.setFlag(substract|halfcarry, true)
	})
}

func scf() opCode {
	return opCodeFn(func(c *CPU, s *ocState) {
		c.setFlag(carry, true)
		c.setFlag(substract|halfcarry, false)
	})
}

func ccf() opCode {
	return opCodeFn(func(c *CPU, s *ocState) {
		c.setFlag(carry, !c.hasFlag(carry))
		c.setFlag(substract|halfcarry, false)
	})
}

func sla(r readerwriter) opCode {
	return pipe(
		r.Read(),
		opCodeFn(func(c *CPU, state *ocState) {
			state.pushB(c.sla(state.popB()))
		}),
		r.Write(),
	)
}

func sra(r readerwriter) opCode {
	return pipe(
		r.Read(),
		opCodeFn(func(c *CPU, state *ocState) {
			state.pushB(c.sra(state.popB()))
		}),
		r.Write(),
	)
}

func srl(r readerwriter) opCode {
	return pipe(
		r.Read(),
		opCodeFn(func(c *CPU, state *ocState) {
			state.pushB(c.srl(state.popB()))
		}),
		r.Write(),
	)
}

func swap(r readerwriter) opCode {
	return pipe(
		r.Read(),
		opCodeFn(func(c *CPU, state *ocState) {
			state.pushB(c.swap(state.popB()))
		}),
		r.Write(),
	)
}

func bit(n byte, r reader) opCode {
	return pipe(
		r.Read(),
		opCodeFn(func(c *CPU, state *ocState) {
			val := state.popB()
			c.testBit(n, val)
		}),
	)
}

func res(n byte, r readerwriter) opCode {
	return pipe(
		r.Read(),
		opCodeFn(func(c *CPU, state *ocState) {
			state.pushB(c.resetBit(n, state.popB()))
		}),
		r.Write(),
	)
}

func set(n byte, r readerwriter) opCode {
	return pipe(
		r.Read(),
		opCodeFn(func(c *CPU, state *ocState) {
			state.pushB(c.setBit(n, state.popB()))
		}),
		r.Write(),
	)
}

func pop(r reg16) opCode {
	incSP := opCodeFn(func(c *CPU, s *ocState) {
		c.sp++
	})

	return pipe(
		sp.Read(),
		readByte{},
		incSP,
		sp.Read(),
		readByte{},
		incSP,
		r.Write(),
	)
}

func push(r reg16) opCode {
	return pipe(
		r.Read(),
		opCodeFn(func(c *CPU, s *ocState) {
			c.sp -= 2
			s.pushW(c.sp)
		}),
		delay{},
		writeWord(),
	)
}

func val(b byte) opCode {
	return opCodeFn(func(c *CPU, s *ocState) {
		s.pushB(b)
	})
}

func ei() opCode {
	return opCodeFn(func(c *CPU, s *ocState) {
		c.imeScheduled = true
	})
}

func di() opCode {
	return opCodeFn(func(c *CPU, s *ocState) {
		c.ime = false
		c.imeScheduled = false
	})
}

func addSP() opCode {
	return pipe(
		paramB(),
		opCodeFn(func(c *CPU, s *ocState) {
			n := s.popB()
			result := uint16(int32(c.sp) + int32(int8(n)))

			check := uint16(c.sp ^ uint16(n) ^ ((c.sp + uint16(n)) & 0xffff))

			c.setFlag(carry, (check&0x100) == 0x100)
			c.setFlag(halfcarry, (check&0x10) == 0x10)
			c.setFlag(zero|substract, false)
			c.sp = result
		}),
		delay{}, delay{},
	)
}

func ldHLSPn() opCode {
	return pipe(
		paramB(),
		opCodeFn(func(c *CPU, s *ocState) {
			n := int8(s.popB())
			HL := uint16(int32(c.sp) + int32(n))

			check := uint16(c.sp ^ uint16(n) ^ ((c.sp + uint16(n)) & 0xffff))

			c.setFlag(carry, (check&0x100) == 0x100)
			c.setFlag(halfcarry, (check&0x10) == 0x10)
			c.setFlag(zero|substract, false)

			c.h, c.l = bytes(uint16(HL))
		}),
		delay{},
	)
}
