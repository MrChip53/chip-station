package utilities

type Stack struct {
	stack   []uint16
	pointer int
}

func NewStack(size int) *Stack {
	return &Stack{
		stack:   make([]uint16, size),
		pointer: 0,
	}
}

func (s *Stack) Push(value uint16) {
	if s.pointer >= len(s.stack) {
		panic("stack overflow")
	}

	s.stack[s.pointer] = value
	s.pointer++
}

func (s *Stack) Pop() uint16 {
	if s.pointer == 0 {
		panic("stack underflow")
	}

	s.pointer--
	return s.stack[s.pointer]
}
