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

/// Ring buffer based

type RQueue[T any] struct {
	buf        []T
	head, tail int
	Len        int
}

func (q *RQueue[T]) Enqueue(v T) {
	if q.Len == len(q.buf) {
		q.grow()
	}
	q.buf[q.tail] = v
	q.tail = (q.tail + 1) % len(q.buf)
	q.Len++
}

func (q *RQueue[T]) Peek() T {
	return q.buf[q.head]
}

func (q *RQueue[T]) Dequeue() (T, bool) {
	var zero T
	if q.Len == 0 {
		return zero, false
	}
	v := q.buf[q.head]
	q.buf[q.head] = zero
	q.head = (q.head + 1) % len(q.buf)
	q.Len--
	return v, true
}

func (q *RQueue[T]) grow() {
	newCap := max(8, len(q.buf)*2)
	newBuf := make([]T, newCap)
	if q.head < q.tail {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		head := copy(newBuf, q.buf[q.head:])
		copy(newBuf[head:], q.buf[:q.tail])
	}
	q.head = 0
	q.tail = q.Len
	q.buf = newBuf
}
