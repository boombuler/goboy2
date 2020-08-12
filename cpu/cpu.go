package cpu

import (
	"fmt"

	"github.com/boombuler/goboy2/consts"
	"github.com/boombuler/goboy2/mmu"
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

	key1         *key1Reg
	ime          bool
	imeScheduled bool
	haltEnabled  bool
	err          error
	haltBug      bool

	curOpCode   opCode
	opCodeState *ocState
	rootOC      opCode

	OnExecOpCode func(opCode string)

	Dump bool
}

// New returns a new cpu connected with the given mmu
func New(mmu mmu.MMU) *CPU {
	cpu := &CPU{
		mmu:         mmu,
		opCodeState: newState(),
		rootOC:      nextOpCode(),
	}
	if mmu.GBC() {
		cpu.key1 = new(key1Reg)
		mmu.AddIODevice(cpu.key1, consts.AddrKEY1)
	}
	return cpu
}

func (c *CPU) Init(noBoot bool) {
	if noBoot {
		if c.mmu.GBC() {
			c.SetRegisterValues(0x0100, 0xFFFE, 0x11, 0x00, 0x00, 0x00, 0x08, 0x80, 0x00, 0x7C)
		} else {
			c.SetRegisterValues(0x0100, 0xFFFE, 0x01, 0x00, 0x13, 0x00, 0xD8, 0xB0, 0x01, 0x4D)
		}
	}
}

// DoubleSpeed checks if the cpu is running in double speed mode.
func (cpu *CPU) DoubleSpeed() bool {
	return cpu.key1 != nil && cpu.key1.dblSpeed
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
	if info, ok := oc.(labeledOpCode); ok {
		if cpu.Dump {
			fmt.Printf("%-15s PC: 0x%04X  SP: 0x%04X  A: 0x%02X  B: 0x%02X  C: 0x%02X  D: 0x%02X  E: 0x%02X  H: 0x%02X  L: 0x%02X  %v\n", info.Label(), cpu.pc, cpu.sp, cpu.a, cpu.b, cpu.c, cpu.d, cpu.e, cpu.h, cpu.l, cpu.f)
		}
		if fn := cpu.OnExecOpCode; fn != nil {
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

		scheduled := cpu.imeScheduled
		cpu.setOPCode(cpu.rootOC)

		if scheduled {
			cpu.ime = true
			cpu.imeScheduled = false
		}

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
