package algos

type SNode[T any] struct {
	v    T
	prev *SNode[T]
}

type Stack[T any] struct {
	Len  int
	head *SNode[T]
}

func (s *Stack[T]) Push(item T) {
	node := &SNode[T]{v: item, prev: s.head}
	s.head = node
	s.Len++
}

func (s *Stack[T]) Peek() T {
	return s.head.v
}

func (s *Stack[T]) Pop() (val T, ok bool) {
	if s.Len == 0 {
		var t T
		return t, false
	}
	res := s.head.v
	s.head = s.head.prev
	s.Len--
	return res, true
}
