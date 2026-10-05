package notify

import (
	"context"
	"sync"
	"testing"
	"time"
)

type countingTransport struct {
	mu    sync.Mutex
	count int
}

func (c *countingTransport) Deliver(ctx context.Context, to, body string) error {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
	return nil
}

func TestDrainDelivers(t *testing.T) {
	tr := &countingTransport{}
	w := &Worker{Queue: &Queue{}, Mailer: &Mailer{Transport: tr}, Timeout: time.Second}
	w.Queue.Enqueue("a@example.com", "hello")
	w.Drain()
	if tr.count != 1 {
		t.Fatalf("deliveries = %d, want 1", tr.count)
	}
}
