package controller

import (
	"math/rand/v2"
	"sync"
)

// WorkTagAllocatorOption configures a workTagAllocator.
type WorkTagAllocatorOption func(*workTagAllocator)

// WithRandomN sets the random index function for the allocator.
// Used in tests to produce deterministic work tag sequences.
func WithRandomN(fn func(int) int) WorkTagAllocatorOption {
	return func(a *workTagAllocator) { a.nth = fn }
}

// workTagAllocator allocates work tags from a rotating pool,
// guaranteeing the same tag is never returned twice in a row.
type workTagAllocator struct {
	mu         sync.Mutex
	selectable []string
	consumed   []string
	nth        func(int) int
}

func newWorkTagAllocator(tags []string, opts ...WorkTagAllocatorOption) *workTagAllocator {
	if len(tags) == 0 {
		panic("workTagAllocator: tag list must not be empty")
	}
	selectable := make([]string, len(tags))
	copy(selectable, tags)

	a := &workTagAllocator{
		selectable: selectable,
		nth:        rand.IntN,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *workTagAllocator) Allocate() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.selectable) == 0 {
		n := len(a.consumed)
		if n == 1 {
			a.selectable = a.consumed
			a.consumed = nil
		} else {
			a.selectable = make([]string, n-1)
			copy(a.selectable, a.consumed[:n-1])
			a.consumed = []string{a.consumed[n-1]}
		}
	}
	idx := a.nth(len(a.selectable))
	tag := a.selectable[idx]
	a.selectable = append(a.selectable[:idx], a.selectable[idx+1:]...)
	a.consumed = append(a.consumed, tag)
	return tag
}
