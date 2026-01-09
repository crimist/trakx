package pool

import (
	"sync"
	"sync/atomic"
)

// Pool wraps sync.Pool with optional reset behavior and a created counter.
type Pool[T any] struct {
	pool    sync.Pool
	reset   func(T) T
	created atomic.Int64
}

// New returns a new pool that creates values via newFn and resets via resetFn.
func New[T any](newFn func() T, resetFn func(T) T) *Pool[T] {
	p := &Pool[T]{reset: resetFn}
	p.pool.New = func() any {
		p.created.Add(1)
		return newFn()
	}
	return p
}

// Get returns a value from the pool.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put resets and returns a value to the pool.
func (p *Pool[T]) Put(v T) {
	if p.reset != nil {
		v = p.reset(v)
	}
	p.pool.Put(v)
}

// Created returns the total number of values allocated by the pool.
func (p *Pool[T]) Created() int64 {
	return p.created.Load()
}
