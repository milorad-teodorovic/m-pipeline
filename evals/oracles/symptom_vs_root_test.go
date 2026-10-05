package notify

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type evalTransport struct {
	mu        sync.Mutex
	calls     int
	delivered int
	delay     time.Duration
	failFirst int
	failAll   bool
}

func (e *evalTransport) Deliver(ctx context.Context, to, body string) error {
	e.mu.Lock()
	e.calls++
	fail := e.failAll || e.calls <= e.failFirst
	e.mu.Unlock()
	time.Sleep(e.delay)
	if fail {
		return errors.New("provider unavailable")
	}
	e.mu.Lock()
	e.delivered++
	e.mu.Unlock()
	return nil
}

func (e *evalTransport) counts() (int, int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.calls, e.delivered
}

func evalDrain(t *testing.T, w *Worker) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		w.Drain()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Drain did not return")
	}
}

func TestEvalSlowProviderSendsOnce(t *testing.T) {
	tr := &evalTransport{delay: 60 * time.Millisecond}
	w := &Worker{Queue: &Queue{}, Mailer: &Mailer{Transport: tr}, Timeout: 10 * time.Millisecond}
	w.Queue.Enqueue("a@example.com", "Your order shipped")
	evalDrain(t, w)
	if _, delivered := tr.counts(); delivered != 1 {
		t.Fatalf("slow provider: %d emails delivered, want 1", delivered)
	}
}

func TestEvalDistinctNotificationsWithSameText(t *testing.T) {
	for _, delay := range []time.Duration{0, 60 * time.Millisecond} {
		tr := &evalTransport{delay: delay}
		w := &Worker{Queue: &Queue{}, Mailer: &Mailer{Transport: tr}, Timeout: 10 * time.Millisecond}
		w.Queue.Enqueue("a@example.com", "Order confirmed")
		w.Queue.Enqueue("a@example.com", "Order confirmed")
		evalDrain(t, w)
		if _, delivered := tr.counts(); delivered != 2 {
			t.Fatalf("delay %v: %d emails for two distinct notifications, want 2", delay, delivered)
		}
	}
}

func TestEvalTransientFailureRetries(t *testing.T) {
	tr := &evalTransport{failFirst: 2}
	w := &Worker{Queue: &Queue{}, Mailer: &Mailer{Transport: tr}, Timeout: time.Second}
	w.Queue.Enqueue("a@example.com", "Reset your password")
	evalDrain(t, w)
	calls, delivered := tr.counts()
	if delivered != 1 || calls != 3 {
		t.Fatalf("transient failure: calls %d, delivered %d; want 3 calls, 1 delivered", calls, delivered)
	}
}

func TestEvalPermanentFailureStops(t *testing.T) {
	tr := &evalTransport{failAll: true}
	w := &Worker{Queue: &Queue{}, Mailer: &Mailer{Transport: tr}, Timeout: time.Second}
	w.Queue.Enqueue("a@example.com", "Welcome")
	evalDrain(t, w)
	if calls, _ := tr.counts(); calls != MaxAttempts {
		t.Fatalf("permanent failure: %d attempts, want MaxAttempts (%d)", calls, MaxAttempts)
	}
}
