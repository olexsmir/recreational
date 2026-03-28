package algos

type QNode[T any] struct {
	v    T
	next *QNode[T]
}

type Queue[T any] struct {
	Len  int
	head *QNode[T]
	tail *QNode[T]
}

func (q *Queue[T]) Push(item T) {
	node := &QNode[T]{v: item}
	if q.Len == 0 {
		q.head = node
		q.tail = node
	} else {
		q.tail.next = node
		q.tail = node
	}
	q.Len++
}

func (q *Queue[T]) Peek() T {
	return q.head.v
}

func (q *Queue[T]) Pop() (T, bool) {
	if q.Len == 0 {
		var t T
		return t, false
	}

	res := q.head.v
	q.head = q.head.next
	if q.head == nil {
		q.tail = nil
	}
	q.Len--
	return res, true
}
