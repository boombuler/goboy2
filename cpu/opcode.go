package cpu

type opCode interface {
	Exec(cpu *CPU, state *ocState)
	Next(state *ocState) opCode
	TakesCycle() bool
}

type namedOpCode interface {
	Name() string
}

type named struct {
	name string
	opCode
}

func (n named) Name() string {
	return n.name
}

type ocState struct {
	buf []byte
}

func newState() *ocState {
	return &ocState{
		buf: make([]byte, 0, 0x0f),
	}
}

func (s *ocState) pushB(val byte) {
	s.buf = append(s.buf, val)
}
func (s *ocState) popB() byte {
	d := s.buf[len(s.buf)-1]
	s.buf = s.buf[:len(s.buf)-1]
	return d
}
func (s *ocState) peekB() byte {
	return s.buf[len(s.buf)-1]
}
func (s *ocState) clear() {
	s.buf = s.buf[:0]
}

func (s *ocState) pushW(val uint16) {
	hi, lo := bytes(val)
	s.pushB(lo)
	s.pushB(hi)
}
func (s *ocState) popW() uint16 {
	hi := s.popB()
	lo := s.popB()
	return word(hi, lo)
}
func (s *ocState) peekW() uint16 {
	last := len(s.buf) - 1
	return word(s.buf[last], s.buf[last-1])
}
