package main

import (
	"fmt"
	"sync"
	"time"
)

type Throttler struct {
	Concurrency int
	Quota       int
	Clients     map[string]int
	Requests    map[string]int
	*sync.Mutex
}

func NewThrottler(concurrency int, quota int) *Throttler {
	return &Throttler{
		Concurrency: concurrency,
		Quota:       quota,
		Clients:     make(map[string]int),
		Requests:    make(map[string]int),
		Mutex:       &sync.Mutex{},
	}
}

func (t *Throttler) StartPeriodicFlush() {
	go func() {
		for {
			t.Flush()
			time.Sleep(time.Second * 5)
		}
	}()
}

func (t *Throttler) Add(ip string) error {
	t.Lock()
	defer t.Unlock()

	t.Requests[ip]++

	if t.Requests[ip] > t.Quota || t.Clients[ip] >= t.Concurrency {
		return fmt.Errorf("Too many requests")
	}

	t.Clients[ip]++
	return nil
}

func (t *Throttler) Remove(ip string) {
	t.Lock()
	defer t.Unlock()

	t.Clients[ip]--

	if t.Clients[ip] < 0 {
		t.Clients[ip] = 0
	}
}

func (t *Throttler) Flush() {
	t.Lock()
	defer t.Unlock()

	for k := range t.Clients {
		delete(t.Clients, k)
	}

	for k := range t.Requests {
		delete(t.Requests, k)
	}
}
