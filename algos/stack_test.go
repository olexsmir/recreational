package algos

import "testing"

func TestStack(t *testing.T) {
	s := Stack[int]{}

	s.Push(5)
	s.Push(7)
	s.Push(9)

	v, ok := s.Pop()
	is(t, ok, true)
	is(t, v, 9)

	is(t, s.Len, 2)

	s.Push(11)

	v, ok = s.Pop()
	is(t, ok, true)
	is(t, v, 11)

	v, ok = s.Pop()
	is(t, ok, true)
	is(t, v, 7)

	is(t, s.Peek(), 5)

	v, ok = s.Pop()
	is(t, ok, true)
	is(t, v, 5)

	_, ok = s.Pop()
	is(t, ok, false)

	s.Push(69420)
	is(t, s.Peek(), 69420)

	v, ok = s.Pop()
	is(t, ok, true)
	is(t, v, 69420)
}
