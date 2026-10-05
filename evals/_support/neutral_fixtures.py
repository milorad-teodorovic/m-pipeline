"""Plain repositories for the neutral with-and-without-plugin cases.

A neutral case sends an ordinary delivery request that names no /seamark command, so
the with-plugin arm and the no-plugin arm receive the same prompt and the same
repository. Hidden oracle tests in evals/oracles score both arms alike.
"""

from pathlib import Path

MEMORY = {
    ".gitignore": ".checks/\n",
    ".seamark/INDEX.md": "# Fixture project\nGo, standard library only. Test: `go test ./...`.\n",
    ".seamark/STACK.md": "# Stack\nGo standard library.\n",
    ".seamark/PATTERNS.md": "# Patterns\nPure functions; table-driven Go tests.\n",
    ".seamark/TASKS.md": "# Tasks\nThe current request is the only task.\n",
    ".seamark/PROGRESS.md": "# Progress\nFixture initialized; current task has not run.\n",
    ".seamark/GAPS.md": "# Gaps\nNone recorded.\n",
    ".seamark/pipeline.yml": "second_engine:\n  provider: none\n",
}

EMAIL = {
    "go.mod": "module fixture\n\ngo 1.22\n",
    "email.go": '''package email

// NormalizeEmail normalizes an email address for comparison.
func NormalizeEmail(s string) string {
	return s
}
''',
    "email_test.go": '''package email

import "testing"

func TestEmptyEmail(t *testing.T) {
	if got := NormalizeEmail(""); got != "" {
		t.Fatalf("NormalizeEmail(empty) = %q", got)
	}
}
''',
}

CENTS = {
    "go.mod": "module fixture\n\ngo 1.22\n",
    "invoice.go": '''package billing

// Line is one invoice line with its unit price in cents.
type Line struct {
	Description string
	UnitCents   int64
	Quantity    int64
}

// TotalCents returns the sum of the line totals in cents.
func TotalCents(lines []Line) int64 {
	var total int64
	for _, l := range lines {
		total += l.UnitCents * l.Quantity
	}
	return total
}
''',
    "invoice_test.go": '''package billing

import "testing"

func TestTotalCents(t *testing.T) {
	lines := []Line{{"a", 250, 2}, {"b", 1999, 1}}
	if got := TotalCents(lines); got != 2499 {
		t.Fatalf("TotalCents = %d, want 2499", got)
	}
}
''',
}

PAGINATE = {
    "go.mod": "module fixture\n\ngo 1.22\n",
    "paginate.go": '''package catalog

// Page returns the items on the given 1-based page when each page holds size items.
func Page(items []string, page, size int) []string {
	start := (page-1)*size - 1
	if start < 0 {
		start = 0
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}
''',
    "paginate_test.go": '''package catalog

import (
	"reflect"
	"testing"
)

func TestFirstPage(t *testing.T) {
	got := Page([]string{"a", "b", "c", "d", "e"}, 1, 2)
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("Page(1, 2) = %v", got)
	}
}
''',
}

LIMITER = {
    "go.mod": "module fixture\n\ngo 1.22\n",
    "limiter.go": '''package ratelimit

import "time"

// Limiter decides whether a request for a key is allowed.
type Limiter struct {
	limit  int
	window time.Duration
	now    func() time.Time
}

// NewLimiter returns a Limiter that allows limit requests per key in each window.
func NewLimiter(limit int, window time.Duration, now func() time.Time) *Limiter {
	return &Limiter{limit: limit, window: window, now: now}
}

// Allow reports whether a request for key is allowed now.
func (l *Limiter) Allow(key string) bool {
	return true
}
''',
    "limiter_test.go": '''package ratelimit

import (
	"testing"
	"time"
)

func TestFirstRequestAllowed(t *testing.T) {
	l := NewLimiter(1, time.Minute, time.Now)
	if !l.Allow("a") {
		t.Fatal("first request denied")
	}
}
''',
}

REFUND = {
    'errors.go': 'package billing\n\nimport "errors"\n\nvar (\n\t// ErrNotFound reports that no invoice has the requested ID.\n\tErrNotFound = errors.New("invoice not found")\n\t// ErrInvalidAmount reports an amount of zero or less.\n\tErrInvalidAmount = errors.New("amount must be positive")\n\t// ErrOverpayment reports a payment larger than the open balance.\n\tErrOverpayment = errors.New("payment exceeds open balance")\n\t// ErrInvoiceVoid reports an operation on a void invoice.\n\tErrInvoiceVoid = errors.New("invoice is void")\n)\n',
    'events.go': 'package billing\n\nimport "errors"\n\nvar errUnbalanced = errors.New("ledger entries do not balance")\n\n// Event types sent to the Notifier.\nconst (\n\tEventPaid = "invoice.paid"\n)\n\n// Event tells listeners that an invoice changed.\ntype Event struct {\n\tType      string\n\tInvoiceID string\n\tCents     int64\n}\n\n// Notifier receives invoice events. Support tooling and the customer email\n// service listen for every money movement.\ntype Notifier interface {\n\tNotify(Event)\n}\n',
    'go.mod': 'module fixture\n\ngo 1.22\n',
    'invoice.go': 'package billing\n\n// Status is the lifecycle state of an invoice.\ntype Status string\n\nconst (\n\tStatusOpen Status = "open"\n\tStatusPaid Status = "paid"\n\tStatusVoid Status = "void"\n)\n\n// Invoice is a customer invoice. All amounts are integer cents.\ntype Invoice struct {\n\tID            string\n\tTotalCents    int64\n\tPaidCents     int64\n\tRefundedCents int64\n\tStatus        Status\n}\n',
    'ledger.go': 'package billing\n\nimport "sync"\n\n// Entry is one side of a money movement.\ntype Entry struct {\n\tAccount   string\n\tInvoiceID string\n\tCents     int64\n}\n\n// Ledger is the double-entry record of money movements. Every movement posts\n// entries whose Cents sum to zero, so the ledger total is always zero.\ntype Ledger struct {\n\tmu      sync.Mutex\n\tentries []Entry\n}\n\n// Post records the entries of one movement. It rejects a movement whose\n// entries do not sum to zero.\nfunc (l *Ledger) Post(entries ...Entry) error {\n\tvar sum int64\n\tfor _, e := range entries {\n\t\tsum += e.Cents\n\t}\n\tif sum != 0 {\n\t\treturn errUnbalanced\n\t}\n\tl.mu.Lock()\n\tdefer l.mu.Unlock()\n\tl.entries = append(l.entries, entries...)\n\treturn nil\n}\n\n// Balance returns the sum of the entries posted to account.\nfunc (l *Ledger) Balance(account string) int64 {\n\tl.mu.Lock()\n\tdefer l.mu.Unlock()\n\tvar total int64\n\tfor _, e := range l.entries {\n\t\tif e.Account == account {\n\t\t\ttotal += e.Cents\n\t\t}\n\t}\n\treturn total\n}\n\n// Entries returns a copy of every posted entry.\nfunc (l *Ledger) Entries() []Entry {\n\tl.mu.Lock()\n\tdefer l.mu.Unlock()\n\treturn append([]Entry(nil), l.entries...)\n}\n',
    'service.go': 'package billing\n\n// Account names used in ledger entries.\nconst (\n\tAccountCash    = "cash"\n\tAccountRevenue = "revenue"\n)\n\n// Service applies money movements to invoices.\ntype Service struct {\n\tStore    *Store\n\tLedger   *Ledger\n\tNotifier Notifier\n}\n\n// Pay records a payment of cents against the invoice with the given ID.\n// It returns ErrInvalidAmount when cents <= 0, ErrNotFound for an unknown ID,\n// ErrInvoiceVoid for a void invoice, and ErrOverpayment when cents exceeds the\n// open balance. A successful payment posts cash and revenue entries and sends\n// an EventPaid event.\nfunc (s *Service) Pay(id string, cents int64) error {\n\tif cents <= 0 {\n\t\treturn ErrInvalidAmount\n\t}\n\terr := s.Store.Update(id, func(inv *Invoice) error {\n\t\tif inv.Status == StatusVoid {\n\t\t\treturn ErrInvoiceVoid\n\t\t}\n\t\tif cents > inv.TotalCents-inv.PaidCents {\n\t\t\treturn ErrOverpayment\n\t\t}\n\t\tinv.PaidCents += cents\n\t\tif inv.PaidCents == inv.TotalCents {\n\t\t\tinv.Status = StatusPaid\n\t\t}\n\t\treturn nil\n\t})\n\tif err != nil {\n\t\treturn err\n\t}\n\tif err := s.Ledger.Post(\n\t\tEntry{Account: AccountCash, InvoiceID: id, Cents: cents},\n\t\tEntry{Account: AccountRevenue, InvoiceID: id, Cents: -cents},\n\t); err != nil {\n\t\treturn err\n\t}\n\ts.Notifier.Notify(Event{Type: EventPaid, InvoiceID: id, Cents: cents})\n\treturn nil\n}\n',
    'service_test.go': 'package billing\n\nimport (\n\t"errors"\n\t"testing"\n)\n\ntype recorder struct{ events []Event }\n\nfunc (r *recorder) Notify(e Event) { r.events = append(r.events, e) }\n\nfunc TestPay(t *testing.T) {\n\trec := &recorder{}\n\ts := &Service{Store: NewStore(Invoice{ID: "i1", TotalCents: 1000, Status: StatusOpen}), Ledger: &Ledger{}, Notifier: rec}\n\tif err := s.Pay("i1", 1000); err != nil {\n\t\tt.Fatal(err)\n\t}\n\tif err := s.Pay("i1", 1); !errors.Is(err, ErrOverpayment) {\n\t\tt.Fatalf("overpayment err = %v", err)\n\t}\n\tif got := s.Ledger.Balance(AccountCash); got != 1000 {\n\t\tt.Fatalf("cash = %d", got)\n\t}\n\tif len(rec.events) != 1 || rec.events[0].Type != EventPaid {\n\t\tt.Fatalf("events = %v", rec.events)\n\t}\n}\n',
    'store.go': 'package billing\n\nimport (\n\t"sync"\n\t"time"\n)\n\n// roundTrip is the latency of one call to the invoice database.\nconst roundTrip = time.Millisecond\n\n// Store is the client for the invoice database. Each call is one database\n// round trip. The database serializes every change.\ntype Store struct {\n\tmu       sync.Mutex\n\tinvoices map[string]Invoice\n}\n\n// NewStore returns a Store that holds the given invoices.\nfunc NewStore(invoices ...Invoice) *Store {\n\ts := &Store{invoices: map[string]Invoice{}}\n\tfor _, inv := range invoices {\n\t\ts.invoices[inv.ID] = inv\n\t}\n\treturn s\n}\n\n// Get returns a copy of the invoice with the given ID, or ErrNotFound.\nfunc (s *Store) Get(id string) (Invoice, error) {\n\ttime.Sleep(roundTrip)\n\ts.mu.Lock()\n\tdefer s.mu.Unlock()\n\tinv, ok := s.invoices[id]\n\tif !ok {\n\t\treturn Invoice{}, ErrNotFound\n\t}\n\treturn inv, nil\n}\n\n// Update applies fn to the invoice with the given ID while the store is locked.\n// Every check that decides a change must run inside fn, so that two concurrent\n// changes cannot both pass the check. The change is saved only when fn returns nil.\nfunc (s *Store) Update(id string, fn func(*Invoice) error) error {\n\ttime.Sleep(roundTrip)\n\ts.mu.Lock()\n\tdefer s.mu.Unlock()\n\tinv, ok := s.invoices[id]\n\tif !ok {\n\t\treturn ErrNotFound\n\t}\n\tif err := fn(&inv); err != nil {\n\t\treturn err\n\t}\n\ts.invoices[id] = inv\n\treturn nil\n}\n',
}

LOOKUP = {
    'api/profile.go': 'package api\n\nimport (\n\t"fixture/users"\n\t"net/http"\n)\n\n// ProfileHandler serves the display name of the user named by the id query parameter.\nfunc ProfileHandler(s *users.Store) http.Handler {\n\treturn http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {\n\t\tu := s.Find(r.URL.Query().Get("id"))\n\t\tw.Header().Set("Content-Type", "text/plain")\n\t\t_, _ = w.Write([]byte(u.Name))\n\t})\n}\n',
    'go.mod': 'module fixture\n\ngo 1.22\n',
    'users/greeting.go': 'package users\n\n// Greeting returns the banner text for the visitor with the given ID.\n// Visitors without an account are guests.\nfunc Greeting(s *Store, id string) string {\n\tu := s.Find(id)\n\tif u.Name == "" {\n\t\treturn "Hello, guest"\n\t}\n\treturn "Hello, " + u.Name\n}\n',
    'users/permissions.go': 'package users\n\n// Document is an editable document.\ntype Document struct {\n\tID    string\n\tOwner string\n}\n\n// CanEdit reports whether the user with the given ID may edit doc.\n// Admins may edit every document; other users may edit the documents they own.\nfunc CanEdit(s *Store, id string, doc Document) bool {\n\tu := s.Find(id)\n\treturn u.Role == "admin" || doc.Owner == id\n}\n',
    'users/store.go': 'package users\n\n// User is an account holder.\ntype User struct {\n\tID   string\n\tName string\n\tRole string\n}\n\n// Store holds users by ID.\ntype Store struct {\n\tusers map[string]User\n}\n\n// NewStore returns a Store that holds the given users.\nfunc NewStore(users ...User) *Store {\n\ts := &Store{users: map[string]User{}}\n\tfor _, u := range users {\n\t\ts.users[u.ID] = u\n\t}\n\treturn s\n}\n\n// Find returns the user with the given ID, or the zero User when none exists.\nfunc (s *Store) Find(id string) User {\n\treturn s.users[id]\n}\n',
    'users/store_test.go': 'package users\n\nimport "testing"\n\nfunc TestGreeting(t *testing.T) {\n\ts := NewStore(User{ID: "u1", Name: "Ana", Role: "member"})\n\tif got := Greeting(s, "u1"); got != "Hello, Ana" {\n\t\tt.Fatalf("Greeting(u1) = %q", got)\n\t}\n}\n',
}

CACHE = {
    'backend.go': 'package pricing\n\nimport (\n\t"context"\n\t"time"\n)\n\n// Backend is the upstream price service. It is slow and rate-limited: the\n// vendor bills every call and suspends the account when one product is\n// requested more than once at the same time. A failed lookup is transient and\n// must be retried on the next request, never remembered.\ntype Backend interface {\n\tPrice(ctx context.Context, sku string) (int64, error)\n}\n\n// Clock returns the current time. Production code uses time.Now.\ntype Clock func() time.Time\n',
    'go.mod': 'module fixture\n\ngo 1.22\n',
    'handler.go': 'package pricing\n\nimport (\n\t"context"\n\t"strconv"\n)\n\n// Quote returns the display price of sku in cents as a decimal string.\nfunc Quote(ctx context.Context, b Backend, sku string) (string, error) {\n\tcents, err := b.Price(ctx, sku)\n\tif err != nil {\n\t\treturn "", err\n\t}\n\treturn strconv.FormatInt(cents, 10), nil\n}\n',
    'handler_test.go': 'package pricing\n\nimport (\n\t"context"\n\t"testing"\n)\n\ntype fixed int64\n\nfunc (f fixed) Price(context.Context, string) (int64, error) { return int64(f), nil }\n\nfunc TestQuote(t *testing.T) {\n\tgot, err := Quote(context.Background(), fixed(1250), "sku-1")\n\tif err != nil || got != "1250" {\n\t\tt.Fatalf("Quote = %q, %v", got, err)\n\t}\n}\n',
}

NEUTRAL = {
    "develop--neutral-email": EMAIL,
    "develop--neutral-cents": CENTS,
    "develop--neutral-paginate": PAGINATE,
    "develop--neutral-limiter": LIMITER,
    "develop--neutral-refund": REFUND,
    "develop--neutral-lookup": LOOKUP,
    "develop--neutral-cache": CACHE,
}


def neutral_name(name):
    """Return the neutral case whose repository a case uses.

    A pipeline twin "develop--<task>" starts from the same repository as its
    plain twin "develop--neutral-<task>".
    """
    if name.startswith("develop--neutral-"):
        return name
    return "develop--neutral-" + name.split("--", 1)[1] if name.startswith("develop--") else name


def repository(name):
    """Return the file map stored under evals/<case>/repo, or None."""
    root = Path(__file__).resolve().parents[1] / name / "repo"
    if not root.is_dir():
        return None
    return {str(p.relative_to(root)): p.read_text() for p in sorted(root.rglob("*")) if p.is_file()}


def prepare_neutral(name):
    """Return the committed file map for a neutral case or its twin, or None for other cases."""
    source = neutral_name(name)
    files = NEUTRAL.get(source) or repository(source)
    if files is None:
        return None
    return {**MEMORY, **files}
