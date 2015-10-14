package main

import (
	"fmt"
	"sync"
)

type Throttler struct {
	Quota   int
	Clients map[string]int
	*sync.Mutex
}

func NewThrottler(quota int) *Throttler {
	return &Throttler{
		Quota:   quota,
		Clients: make(map[string]int),
		Mutex:   &sync.Mutex{},
	}
}

func (t *Throttler) Add(ip string) error {
	t.Lock()
	defer t.Unlock()

	// Bypass all requests if quota is not set
	if t.Quota < 0 {
		return nil
	}

	if t.Clients[ip] >= t.Quota {
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
