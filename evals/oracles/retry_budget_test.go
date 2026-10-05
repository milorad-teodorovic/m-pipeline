package partner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type evalStep struct {
	status int
	header map[string]string
}

type evalServer struct {
	mu     sync.Mutex
	steps  []evalStep
	bodies []string
	srv    *httptest.Server
}

func newEvalServer(t *testing.T, steps ...evalStep) *evalServer {
	t.Helper()
	s := &evalServer{steps: steps}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.mu.Lock()
		i := len(s.bodies)
		s.bodies = append(s.bodies, string(body))
		s.mu.Unlock()
		step := evalStep{status: http.StatusOK}
		if i < len(s.steps) {
			step = s.steps[i]
		} else if len(s.steps) > 0 {
			step = s.steps[len(s.steps)-1]
		}
		for k, v := range step.header {
			w.Header().Set(k, v)
		}
		w.WriteHeader(step.status)
		_, _ = io.WriteString(w, fmt.Sprintf("attempt %d", i+1))
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *evalServer) attempts() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.bodies)
}

type evalCloseTracker struct {
	base    http.RoundTripper
	mu      sync.Mutex
	opened  int
	closed  int
	failFor int
}

type evalBody struct {
	io.ReadCloser
	t    *evalCloseTracker
	once sync.Once
}

func (b *evalBody) Close() error {
	b.once.Do(func() {
		b.t.mu.Lock()
		b.t.closed++
		b.t.mu.Unlock()
	})
	return b.ReadCloser.Close()
}

func (t *evalCloseTracker) RoundTrip(r *http.Request) (*http.Response, error) {
	t.mu.Lock()
	if t.failFor > 0 {
		t.failFor--
		t.mu.Unlock()
		return nil, errors.New("connection reset by peer")
	}
	t.mu.Unlock()
	resp, err := t.base.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	t.mu.Lock()
	t.opened++
	t.mu.Unlock()
	resp.Body = &evalBody{ReadCloser: resp.Body, t: t}
	return resp, nil
}

type evalHarness struct {
	client  *Client
	tracker *evalCloseTracker
	mu      sync.Mutex
	waits   []time.Duration
}

func newEvalHarness(now time.Time) *evalHarness {
	h := &evalHarness{tracker: &evalCloseTracker{base: http.DefaultTransport}}
	h.client = &Client{
		HTTP: &http.Client{Transport: h.tracker},
		Sleep: func(ctx context.Context, d time.Duration) error {
			h.mu.Lock()
			h.waits = append(h.waits, d)
			h.mu.Unlock()
			return nil
		},
		Now: func() time.Time { return now },
	}
	return h
}

func (h *evalHarness) recorded() []time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]time.Duration(nil), h.waits...)
}

func evalDo(t *testing.T, h *evalHarness, ctx context.Context, method, url string, body io.Reader, header map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatalf("%s: unexpected error %v", method, err)
	}
	return resp
}

func evalClose(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

var unavailable = evalStep{status: http.StatusServiceUnavailable}

func TestEvalRetriesTransientStatus(t *testing.T) {
	for _, status := range []int{http.StatusServiceUnavailable, http.StatusTooManyRequests} {
		s := newEvalServer(t, evalStep{status: status}, evalStep{status: status}, evalStep{status: http.StatusOK})
		h := newEvalHarness(time.Now())
		resp := evalDo(t, h, context.Background(), http.MethodGet, s.srv.URL, nil, nil)
		evalClose(resp)
		if resp.StatusCode != http.StatusOK || s.attempts() != 3 {
			t.Errorf("status %d: got %d after %d attempts, want 200 after 3", status, resp.StatusCode, s.attempts())
		}
		if got := fmt.Sprint(h.recorded()); got != fmt.Sprint([]time.Duration{200 * time.Millisecond, 400 * time.Millisecond}) {
			t.Errorf("status %d: waits %s, want [200ms 400ms]", status, got)
		}
	}
}

func TestEvalAttemptCapReturnsLastResponse(t *testing.T) {
	s := newEvalServer(t, unavailable)
	h := newEvalHarness(time.Now())
	resp := evalDo(t, h, context.Background(), http.MethodGet, s.srv.URL, nil, nil)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable || s.attempts() != 3 || string(body) != "attempt 3" {
		t.Errorf("got %d %q after %d attempts, want the third 503 response", resp.StatusCode, body, s.attempts())
	}
}

func TestEvalPermanentStatusNotRetried(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusBadRequest} {
		s := newEvalServer(t, evalStep{status: status}, evalStep{status: http.StatusOK})
		h := newEvalHarness(time.Now())
		resp := evalDo(t, h, context.Background(), http.MethodGet, s.srv.URL, nil, nil)
		evalClose(resp)
		if resp.StatusCode != status || s.attempts() != 1 {
			t.Errorf("status %d: got %d after %d attempts, want no retry", status, resp.StatusCode, s.attempts())
		}
	}
}

func TestEvalUnsafeMethods(t *testing.T) {
	cases := []struct {
		method string
		header map[string]string
		want   int
	}{
		{http.MethodPost, nil, 1},
		{http.MethodPatch, nil, 1},
		{http.MethodPost, map[string]string{"Idempotency-Key": "k1"}, 2},
		{http.MethodPut, nil, 2},
		{http.MethodDelete, nil, 2},
	}
	for _, c := range cases {
		s := newEvalServer(t, unavailable, evalStep{status: http.StatusOK})
		h := newEvalHarness(time.Now())
		resp := evalDo(t, h, context.Background(), c.method, s.srv.URL, strings.NewReader("payload"), c.header)
		evalClose(resp)
		if s.attempts() != c.want {
			t.Errorf("%s %v: %d attempts, want %d", c.method, c.header, s.attempts(), c.want)
		}
	}
}

func TestEvalRetryAfter(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	cases := []struct {
		value string
		want  time.Duration
	}{
		{"2", 2 * time.Second},
		{now.Add(3 * time.Second).Format(http.TimeFormat), 3 * time.Second},
		{"soon", 200 * time.Millisecond},
	}
	for _, c := range cases {
		s := newEvalServer(t, evalStep{status: http.StatusServiceUnavailable, header: map[string]string{"Retry-After": c.value}}, evalStep{status: http.StatusOK})
		h := newEvalHarness(now)
		resp := evalDo(t, h, context.Background(), http.MethodGet, s.srv.URL, nil, nil)
		evalClose(resp)
		if got := h.recorded(); len(got) != 1 || got[0] != c.want || s.attempts() != 2 {
			t.Errorf("Retry-After %q: waits %v after %d attempts, want [%s] and 2 attempts", c.value, got, s.attempts(), c.want)
		}
	}
}

func TestEvalDeadlineStopsRetries(t *testing.T) {
	now := time.Now()
	s := newEvalServer(t, evalStep{status: http.StatusServiceUnavailable, header: map[string]string{"Retry-After": "5"}}, evalStep{status: http.StatusOK})
	h := newEvalHarness(now)
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(time.Second))
	defer cancel()
	resp := evalDo(t, h, ctx, http.MethodGet, s.srv.URL, nil, nil)
	evalClose(resp)
	if resp.StatusCode != http.StatusServiceUnavailable || s.attempts() != 1 || len(h.recorded()) != 0 {
		t.Errorf("Retry-After past deadline: got %d after %d attempts and waits %v, want the 503 with no wait", resp.StatusCode, s.attempts(), h.recorded())
	}
	s2 := newEvalServer(t, unavailable)
	h2 := newEvalHarness(now)
	ctx2, cancel2 := context.WithDeadline(context.Background(), now.Add(300*time.Millisecond))
	defer cancel2()
	resp2 := evalDo(t, h2, ctx2, http.MethodGet, s2.srv.URL, nil, nil)
	evalClose(resp2)
	if s2.attempts() != 2 || fmt.Sprint(h2.recorded()) != fmt.Sprint([]time.Duration{200 * time.Millisecond}) {
		t.Errorf("backoff past deadline: %d attempts, waits %v; want 2 attempts and [200ms]", s2.attempts(), h2.recorded())
	}
}

func TestEvalBodyReplay(t *testing.T) {
	s := newEvalServer(t, unavailable, evalStep{status: http.StatusOK})
	h := newEvalHarness(time.Now())
	resp := evalDo(t, h, context.Background(), http.MethodPut, s.srv.URL, strings.NewReader("payload"), nil)
	evalClose(resp)
	s.mu.Lock()
	bodies := append([]string(nil), s.bodies...)
	s.mu.Unlock()
	if fmt.Sprint(bodies) != "[payload payload]" {
		t.Errorf("bodies sent %q, want the same payload twice", bodies)
	}
	s2 := newEvalServer(t, unavailable, evalStep{status: http.StatusOK})
	h2 := newEvalHarness(time.Now())
	req, _ := http.NewRequest(http.MethodPut, s2.srv.URL, io.NopCloser(strings.NewReader("once")))
	req.GetBody = nil
	resp2, err := h2.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	evalClose(resp2)
	if s2.attempts() != 1 || resp2.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("body without GetBody: %d attempts, status %d; want 1 attempt", s2.attempts(), resp2.StatusCode)
	}
}

func TestEvalTransportErrors(t *testing.T) {
	s := newEvalServer(t, evalStep{status: http.StatusOK})
	h := newEvalHarness(time.Now())
	h.tracker.failFor = 1
	resp := evalDo(t, h, context.Background(), http.MethodGet, s.srv.URL, nil, nil)
	evalClose(resp)
	if resp.StatusCode != http.StatusOK || s.attempts() != 1 {
		t.Errorf("GET after transport error: status %d, %d server attempts", resp.StatusCode, s.attempts())
	}
	s2 := newEvalServer(t, evalStep{status: http.StatusOK})
	h2 := newEvalHarness(time.Now())
	h2.tracker.failFor = 1
	req, _ := http.NewRequest(http.MethodPost, s2.srv.URL, strings.NewReader("x"))
	if _, err := h2.client.Do(req); err == nil || s2.attempts() != 0 {
		t.Errorf("POST after transport error: err %v, %d server attempts; want the error and no retry", err, s2.attempts())
	}
}

func TestEvalDiscardedBodiesClosed(t *testing.T) {
	s := newEvalServer(t, unavailable, unavailable, evalStep{status: http.StatusOK})
	h := newEvalHarness(time.Now())
	resp := evalDo(t, h, context.Background(), http.MethodGet, s.srv.URL, nil, nil)
	h.tracker.mu.Lock()
	opened, closed := h.tracker.opened, h.tracker.closed
	h.tracker.mu.Unlock()
	evalClose(resp)
	if opened != 3 || closed != 2 {
		t.Errorf("opened %d responses and closed %d before returning, want 3 and 2", opened, closed)
	}
}
