package util

import "sync"

type ConcurrentQueue[T comparable] struct {
	// array of items
	items []T
	// Mutual exclusion lock
	lock sync.Mutex
}

func (q *ConcurrentQueue[T]) Enqueue(item T) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items = append(q.items, item)
}

func (q *ConcurrentQueue[T]) Dequeue() T {
	q.lock.Lock()
	defer q.lock.Unlock()
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *ConcurrentQueue[T]) Size() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return len(q.items)
}

func (q *ConcurrentQueue[T]) IsEmpty() bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return len(q.items) == 0
}
