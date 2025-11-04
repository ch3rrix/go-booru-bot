package main

import (
	"sync"
	"time"
)

type Debouncer struct {
	mu     sync.Mutex
	timers map[int64]*time.Timer
}

func NewDebouncer() *Debouncer {
	return &Debouncer{
		timers: make(map[int64]*time.Timer),
	}
}

func (d *Debouncer) Do(key int64, delay time.Duration, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if t, exists := d.timers[key]; exists {
		t.Stop()
	}

	d.timers[key] = time.AfterFunc(delay, fn)
}
