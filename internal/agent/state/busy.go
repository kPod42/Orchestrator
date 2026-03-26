package state

import "sync"

type Busy struct {
	mu   sync.Mutex
	busy bool
}

func NewBusy() *Busy {
	return &Busy{}
}

func (b *Busy) Acquire() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.busy {
		return false
	}
	b.busy = true
	return true
}

func (b *Busy) Release() {
	b.mu.Lock()
	b.busy = false
	defer b.mu.Unlock()
}

func (b *Busy) IsBusy() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.busy
}
