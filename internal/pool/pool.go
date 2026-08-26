package pool

type Item interface {
	Reset()
}

type Pool[T Item] struct {
	pool []T
}

func (p *Pool[T]) Get() (T, bool) {
	n := len(p.pool)
	if n == 0 {
		var zero T
		return zero, false
	}
	res := p.pool[n-1]
	p.pool = p.pool[:n-1]

	return res, true
}

func (p *Pool[T]) Put(s T) {
	s.Reset()
	p.pool = append(p.pool, s)
}

func NewPool[T Item]() *Pool[T] {
	pool := Pool[T]{}
	return &pool
}
