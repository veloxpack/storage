package worker

import (
	"github.com/panjf2000/ants/v2"
)

type Pool struct {
	pool *ants.Pool
}

func NewPool(size int) (*Pool, error) {
	opts := ants.Options{Nonblocking: true}
	pool, err := ants.NewPool(size, ants.WithOptions(opts))
	return &Pool{pool}, err
}

func (p *Pool) Submit(task func()) error {
	return p.pool.Submit(task)
}

func (p *Pool) Release() {
	p.pool.Release()
}

func (p *Pool) Running() int {
	return p.pool.Running()
}
