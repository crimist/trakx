package pools

import "sync"

type Pool[T any] struct {
	pool sync.Pool
}

func NewPool[T any]() Pool[T] {
	return Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return new(T)
			},
		},
	}
}

func (pool *Pool[T]) Get() T {
	return pool.pool.Get().(T)
}

func (pool *Pool[T]) Put(data T) {
	pool.pool.Put(data)
}
