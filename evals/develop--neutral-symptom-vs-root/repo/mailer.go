package notify

import (
	"context"
	"sync"
)

// Transport sends one email through the mail provider.
type Transport interface {
	Deliver(ctx context.Context, to, body string) error
}

// Mailer sends emails and delivers each idempotency key at most once.
type Mailer struct {
	Transport Transport
	mu        sync.Mutex
	claimed   map[string]bool
}

// Send delivers body to to unless key was already delivered or is being delivered.
func (m *Mailer) Send(ctx context.Context, key, to, body string) error {
	m.mu.Lock()
	if m.claimed == nil {
		m.claimed = map[string]bool{}
	}
	if m.claimed[key] {
		m.mu.Unlock()
		return nil
	}
	m.claimed[key] = true
	m.mu.Unlock()
	if err := m.Transport.Deliver(ctx, to, body); err != nil {
		m.mu.Lock()
		delete(m.claimed, key)
		m.mu.Unlock()
		return err
	}
	return nil
}
