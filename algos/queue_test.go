package algos

import "testing"

func TestQueue(t *testing.T) {
	q := Queue[int]{}

	q.Push(5)
	q.Push(7)
	q.Push(9)

	v, ok := q.Pop()
	is(t, ok, true)
	is(t, v, 5)

	is(t, q.Len, 2)

	q.Push(11)

	v, ok = q.Pop()
	is(t, ok, true)
	is(t, v, 7)

	v, ok = q.Pop()
	is(t, ok, true)
	is(t, v, 9)

	is(t, q.Peek(), 11)

	v, ok = q.Pop()
	is(t, ok, true)
	is(t, v, 11)

	_, ok = q.Pop()
	is(t, ok, false)

	// everything works after removing all elements
	q.Push(69420)
	is(t, q.Peek(), 69420)

	v, ok = q.Pop()
	is(t, ok, true)
	is(t, v, 69420)
}

func TestRQueue(t *testing.T) {
	q := RQueue[int]{}

	q.Enqueue(5)
	q.Enqueue(7)
	q.Enqueue(9)

	v, ok := q.Dequeue()
	is(t, ok, true)
	is(t, v, 5)

	is(t, q.Len, 2)

	q.Enqueue(11)

	v, ok = q.Dequeue()
	is(t, ok, true)
	is(t, v, 7)

	v, ok = q.Dequeue()
	is(t, ok, true)
	is(t, v, 9)

	is(t, q.Peek(), 11)

	v, ok = q.Dequeue()
	is(t, ok, true)
	is(t, v, 11)

	_, ok = q.Dequeue()
	is(t, ok, false)

	// everything works after removing all elements
	q.Enqueue(69420)
	is(t, q.Peek(), 69420)

	v, ok = q.Dequeue()
	is(t, ok, true)
	is(t, v, 69420)
}

func BenchmarkQueue(b *testing.B) {
	q := &Queue[int]{}
	for range 256 {
		q.Push(1)
	}

	for b.Loop() {
		q.Push(1)
		q.Pop()
	}
}

func BenchmarkRQueue(b *testing.B) {
	q := &RQueue[int]{}
	for range 256 {
		q.Enqueue(1)
	}

	for b.Loop() {
		q.Enqueue(1)
		q.Dequeue()
	}
}

func is[T comparable](tb testing.TB, a, b T) {
	tb.Helper()
	if a != b {
		tb.Fatalf("%+v and %+v are not equal\n", a, b)
	}
}
