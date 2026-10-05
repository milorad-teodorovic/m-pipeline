package ratelimit

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type evalClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *evalClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *evalClock) advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func TestEvalLimiterWindow(t *testing.T) {
	clock := &evalClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	l := NewLimiter(3, time.Minute, clock.now)
	for i := 0; i < 3; i++ {
		if !l.Allow("a") {
			t.Fatalf("request %d denied within limit", i+1)
		}
	}
	if l.Allow("a") {
		t.Fatal("fourth request allowed within one window")
	}
	if !l.Allow("b") {
		t.Fatal("another key shared the first key's budget")
	}
	clock.advance(time.Minute)
	if !l.Allow("a") {
		t.Fatal("request denied after the window elapsed")
	}
}

func TestEvalLimiterConcurrent(t *testing.T) {
	clock := &evalClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	l := NewLimiter(50, time.Minute, clock.now)
	var allowed int64
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow("shared") {
				atomic.AddInt64(&allowed, 1)
			}
		}()
	}
	wg.Wait()
	if allowed != 50 {
		t.Fatalf("allowed %d concurrent requests, want 50", allowed)
	}
}
