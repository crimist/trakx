package pools

import (
	"sync"
	"sync/atomic"
)

type Pool[T any] struct {
	pool    sync.Pool
	reset   func(T)
	created atomic.Int32 // total objects created
}

func NewPool[T any](new func() any, reset func(T)) *Pool[T] {
	pool := Pool[T]{
		reset: reset,
	}

	pool.pool = sync.Pool{
		New: func() any {
			pool.created.Add(1)
			return new()
		},
	}

	return &pool
}

func (pool *Pool[T]) Get() T {
	return pool.pool.Get().(T)
}

func (pool *Pool[T]) Put(data T) {
	if pool.reset != nil {
		pool.reset(data)
	}
	pool.pool.Put(data)
}

func (pool *Pool[T]) Created() int32 {
	return pool.created.Load()
}
