package billing

import (
	"errors"
	"sync"
	"testing"
)

type evalRecorder struct {
	mu     sync.Mutex
	events []Event
}

func (r *evalRecorder) Notify(e Event) {
	r.mu.Lock()
	r.events = append(r.events, e)
	r.mu.Unlock()
}

func (r *evalRecorder) snapshot() []Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Event(nil), r.events...)
}

func evalPaid(t *testing.T, total int64) (*Service, *evalRecorder) {
	t.Helper()
	rec := &evalRecorder{}
	s := &Service{Store: NewStore(Invoice{ID: "i1", TotalCents: total, Status: StatusOpen}, Invoice{ID: "open", TotalCents: 500, Status: StatusOpen}), Ledger: &Ledger{}, Notifier: rec}
	if err := s.Pay("i1", total); err != nil {
		t.Fatal(err)
	}
	return s, rec
}

func evalLedgerTotal(l *Ledger) int64 {
	var sum int64
	for _, e := range l.Entries() {
		sum += e.Cents
	}
	return sum
}

func TestEvalRefundLimits(t *testing.T) {
	s, _ := evalPaid(t, 1000)
	if err := s.Refund("i1", 300); err != nil {
		t.Fatalf("partial refund: %v", err)
	}
	if err := s.Refund("i1", 800); err == nil {
		t.Fatal("refund above the remaining paid amount succeeded")
	}
	if err := s.Refund("i1", 700); err != nil {
		t.Fatalf("refund of the remaining amount: %v", err)
	}
	if err := s.Refund("i1", 1); err == nil {
		t.Fatal("refund after a full refund succeeded")
	}
	if err := s.Refund("open", 1); err == nil {
		t.Fatal("refund of an unpaid invoice succeeded")
	}
}

func TestEvalRefundErrors(t *testing.T) {
	s, _ := evalPaid(t, 1000)
	for _, cents := range []int64{0, -5} {
		if err := s.Refund("i1", cents); !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("Refund(%d) err = %v, want ErrInvalidAmount", cents, err)
		}
	}
	if err := s.Refund("missing", 10); !errors.Is(err, ErrNotFound) {
		t.Errorf("Refund(missing) err = %v, want ErrNotFound", err)
	}
}

func TestEvalRefundLedgerAndEvents(t *testing.T) {
	s, rec := evalPaid(t, 1000)
	before := len(s.Ledger.Entries())
	if err := s.Refund("i1", 5000); err == nil {
		t.Fatal("over-refund succeeded")
	}
	if len(s.Ledger.Entries()) != before || len(rec.snapshot()) != 1 {
		t.Fatal("a rejected refund posted ledger entries or sent an event")
	}
	if err := s.Refund("i1", 300); err != nil {
		t.Fatal(err)
	}
	if got := s.Ledger.Balance(AccountCash); got != 700 {
		t.Errorf("cash balance = %d, want 700", got)
	}
	if got := evalLedgerTotal(s.Ledger); got != 0 {
		t.Errorf("ledger total = %d, want 0", got)
	}
	events := rec.snapshot()
	last := events[len(events)-1]
	if len(events) != 2 || last.InvoiceID != "i1" || last.Cents != 300 || last.Type == EventPaid || last.Type == "" {
		t.Errorf("events after refund = %+v, want one refund event for i1 of 300", events)
	}
}

func TestEvalRefundConcurrent(t *testing.T) {
	s, _ := evalPaid(t, 1000)
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if s.Refund("i1", 30) == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if ok != 33 {
		t.Fatalf("%d concurrent refunds of 30 succeeded against 1000 paid, want 33", ok)
	}
	if got := s.Ledger.Balance(AccountCash); got != 1000-33*30 {
		t.Fatalf("cash balance = %d, want %d", got, 1000-33*30)
	}
}
