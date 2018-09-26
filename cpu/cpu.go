package cpu

import (
	"fmt"
	"goboy2/consts"
	"goboy2/mmu"
)

// SetRegisterValues set all Register values for the cpu
func (cpu *CPU) SetRegisterValues(pc, sp uint16, a, b, c, d, e, f, h, l byte) {
	cpu.pc = pc
	cpu.sp = sp
	cpu.a = a
	cpu.b = b
	cpu.c = c
	cpu.d = d
	cpu.e = e
	cpu.f = flag(f)
	cpu.h = h
	cpu.l = l
}

// GetRegisterValues returns the current register values
func (cpu *CPU) GetRegisterValues() (pc, sp uint16, a, b, c, d, e, f, h, l byte) {
	pc = cpu.pc
	sp = cpu.sp
	a = cpu.a
	b = cpu.b
	c = cpu.c
	d = cpu.d
	e = cpu.e
	f = byte(cpu.f)
	h = cpu.h
	l = cpu.l
	return
}

// CPU is the central processing unit of the gameboy which will consume and execute the program code
type CPU struct {
	mmu mmu.MMU
	registers

	ime         bool
	haltEnabled bool
	err         error
	haltBug     bool

	curOpCode   opCode
	opCodeState *ocState
	rootOC      opCode

	OnExecOpCode func(opCode string)

	Dump bool
}

// New returns a new cpu connected with the given mmu
func New(mmu mmu.MMU) *CPU {
	return &CPU{
		mmu:         mmu,
		opCodeState: newState(),
		rootOC:      nextOpCode(),
	}
}

func (cpu *CPU) setFlag(f flag, val bool) {
	if val {
		cpu.f = cpu.f | f
	} else {
		cpu.f = cpu.f & (0xFF ^ f)
	}
}

func (cpu *CPU) hasFlag(f flag) bool {
	return cpu.f&f == f
}

func (cpu *CPU) nextOpCode(oc opCode, state *ocState) opCode {
	oc = oc.Next(state)
	if fn := cpu.OnExecOpCode; fn != nil {
		if info, ok := oc.(labeledOpCode); ok {
			fn(info.Label())
		}
	}
	return oc
}

func (cpu *CPU) execInstantCodes(oc opCode) opCode {
	for oc != nil && !oc.TakesCycle() {
		oc.Exec(cpu, cpu.opCodeState)
		oc = cpu.nextOpCode(oc, cpu.opCodeState)
	}
	return oc
}

func (cpu *CPU) stepOpCode() {
	oc := cpu.curOpCode

	oc.Exec(cpu, cpu.opCodeState)
	oc = cpu.nextOpCode(oc, cpu.opCodeState)
	cpu.curOpCode = cpu.execInstantCodes(oc)
}

func (cpu *CPU) setOPCode(oc opCode) {
	cpu.curOpCode = cpu.execInstantCodes(oc)
	cpu.stepOpCode()
}

// Step Executes the next cpu step
func (cpu *CPU) Step() {
	if cpu.curOpCode != nil {
		cpu.stepOpCode()
		return
	}

	if !cpu.haltEnabled {
		if cpu.handleInterrupts() {
			return
		}

		cpu.opCodeState.clear()
		if cpu.Dump {
			fmt.Printf("PC: 0x%04X  SP: 0x%04X  A: 0x%02X  B: 0x%02X  C: 0x%02X  D: 0x%02X  E: 0x%02X  H: 0x%02X  L: 0x%02X  %v\n", cpu.pc, cpu.sp, cpu.a, cpu.b, cpu.c, cpu.d, cpu.e, cpu.h, cpu.l, cpu.f)
		}
		cpu.setOPCode(cpu.rootOC)
	} else {
		curIRQFlags := mmu.IRQ(cpu.mmu.Read(consts.AddrIRQEnabled)) & mmu.IRQ(cpu.mmu.Read(consts.AddrIRQFlags))
		if (curIRQFlags & mmu.IRQAll) != mmu.IRQNone {
			cpu.haltEnabled = false
		}
	}
}

func (cpu *CPU) handleInterrupts() bool {
	if cpu.ime {
		curIRQFlags := mmu.IRQ(cpu.mmu.Read(consts.AddrIRQEnabled)) & mmu.IRQ(cpu.mmu.Read(consts.AddrIRQFlags))
		if (curIRQFlags & mmu.IRQAll) != mmu.IRQNone {
			cpu.setOPCode(irqHandlerOpCode)
			return true
		}
	}
	return false
}
