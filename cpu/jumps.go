package cpu

type jump struct {
	f      flag
	neg    bool
	jumpOC opCode
}

func jr(f flag, negate bool) opCode {
	return pipe(
		paramB(),
		&jump{
			f:   f,
			neg: negate,
			jumpOC: pipe(
				delay{},
				opCodeFn(func(c *CPU, s *ocState) {
					n := s.popB()
					c.pc = uint16(int32(c.pc) + int32(int8(n)))
				}),
			),
		},
	)
}

func ret(f flag, negate bool) opCode {
	return pipe(
		delay{},
		&jump{
			f:   f,
			neg: negate,
			jumpOC: pipe(
				pop(pc),
				delay{},
			),
		},
	)
}

func reti() opCode {
	return pipe(
		pop(pc),
		delay{},
		opCodeFn(func(c *CPU, s *ocState) {
			c.ime = true
		}),
	)
}

func jp(f flag, negate bool) opCode {
	return pipe(
		paramW(),
		&jump{
			f:   f,
			neg: negate,
			jumpOC: pipe(
				delay{},
				opCodeFn(func(c *CPU, s *ocState) {
					c.pc = s.popW()
				}),
			),
		},
	)
}

func call(f flag, negate bool) opCode {
	return pipe(
		paramW(),
		&jump{
			f:   f,
			neg: negate,
			jumpOC: pipe(
				push(pc),
				opCodeFn(func(c *CPU, s *ocState) {
					c.pc = s.popW()
				}),
			),
		},
	)
}

func rst(n byte) opCode {
	return pipe(
		push(pc),
		opCodeFn(func(c *CPU, s *ocState) {
			c.pc = uint16(n)
		}),
	)
}

func (j *jump) Exec(cpu *CPU, state *ocState) {
	if cpu.hasFlag(j.f) != j.neg {
		state.pushB(1)
	} else {
		state.pushB(0)
	}
}

func (j *jump) Next(state *ocState) opCode {
	jump := state.popB()
	if jump == 1 {
		return j.jumpOC
	}
	return nil
}
func (j *jump) TakesCycle() bool {
	return false
}
