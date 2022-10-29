package gomodanalysis

import (
	"sync"
	"time"
)

type Backoff struct {
	MaxDelay time.Duration
	MaxSteps int64

	step int64
	mu   sync.Mutex
}

type TimeoutError struct{}

func (e TimeoutError) Error() string {
	return "max retry delay has been reached"
}

func (b *Backoff) LinearDelay() (time.Duration, error) {
	if b.step > b.MaxSteps {
		return 0, TimeoutError{}
	}
	return time.Duration(b.MaxDelay.Nanoseconds() / b.MaxSteps * b.step), nil
}

func (b *Backoff) UpCounter() {
	b.mu.Lock()
	b.step++
	b.mu.Unlock()
}

func (b *Backoff) Reset() {
	b.mu.Lock()
	b.step = 0
	b.mu.Unlock()
}
