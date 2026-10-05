package catalog

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

var evalBase = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

func evalSeed(n int, group int) (*Store, []Product) {
	s := NewStore()
	ids := rand.New(rand.NewSource(7)).Perm(n)
	var all []Product
	for _, id := range ids {
		p := Product{ID: int64(id + 1), Name: fmt.Sprintf("p%d", id+1), CreatedAt: evalBase.Add(time.Duration((id)/group) * time.Minute)}
		s.Add(p)
		all = append(all, p)
	}
	return s, all
}

func evalPage(t *testing.T, s *Store, limit, cursor string) (int, []Product, string, bool) {
	t.Helper()
	q := url.Values{}
	if limit != "" {
		q.Set("limit", limit)
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	rec := httptest.NewRecorder()
	ListHandler(s).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/products?"+q.Encode(), nil))
	if rec.Code != http.StatusOK {
		return rec.Code, nil, "", false
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("response is not a JSON object: %v", err)
	}
	var items []Product
	_ = json.Unmarshal(raw["items"], &items)
	next, has := "", false
	if v, ok := raw["next_cursor"]; ok {
		has = true
		if err := json.Unmarshal(v, &next); err != nil {
			t.Fatalf("next_cursor is not a string: %s", v)
		}
	}
	return rec.Code, items, next, has
}

func evalLess(a, b Product) bool {
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.Before(b.CreatedAt)
	}
	return a.ID < b.ID
}

func evalWalk(t *testing.T, s *Store, limit string, between func(page int, seen []Product)) ([]Product, int) {
	t.Helper()
	var seen []Product
	cursor := ""
	for page := 1; page <= 50; page++ {
		code, items, next, has := evalPage(t, s, limit, cursor)
		if code != http.StatusOK {
			t.Fatalf("page %d: status %d", page, code)
		}
		if !has {
			t.Fatalf("page %d: response has no next_cursor field", page)
		}
		seen = append(seen, items...)
		if next == "" {
			return seen, page
		}
		if len(items) == 0 {
			t.Fatalf("page %d: empty page with a non-empty next_cursor", page)
		}
		if between != nil {
			between(page, seen)
		}
		cursor = next
	}
	t.Fatal("walk did not end within 50 pages")
	return nil, 0
}

func evalCheckOrder(t *testing.T, seen []Product) {
	t.Helper()
	ids := map[int64]bool{}
	for i, p := range seen {
		if ids[p.ID] {
			t.Errorf("item %d returned twice", p.ID)
		}
		ids[p.ID] = true
		if i > 0 && !evalLess(seen[i-1], p) {
			t.Errorf("order broken at %d: %d (%s) after %d (%s)", i, p.ID, p.CreatedAt.Format(time.TimeOnly), seen[i-1].ID, seen[i-1].CreatedAt.Format(time.TimeOnly))
		}
	}
}

func TestEvalStableOrderWithTies(t *testing.T) {
	s, all := evalSeed(25, 5)
	seen, pages := evalWalk(t, s, "7", nil)
	evalCheckOrder(t, seen)
	if len(seen) != len(all) || pages != 4 {
		t.Errorf("walk returned %d items in %d pages, want 25 in 4", len(seen), pages)
	}
}

func TestEvalLastPageExactMultiple(t *testing.T) {
	s, _ := evalSeed(21, 3)
	seen, pages := evalWalk(t, s, "7", nil)
	if len(seen) != 21 || pages != 3 {
		t.Errorf("walk returned %d items in %d pages, want 21 in 3 with an empty next_cursor on page 3", len(seen), pages)
	}
}

func TestEvalCursorSurvivesChanges(t *testing.T) {
	s, all := evalSeed(30, 4)
	deleted := map[int64]bool{}
	seen, _ := evalWalk(t, s, "10", func(page int, seen []Product) {
		if page != 1 {
			return
		}
		last := seen[len(seen)-1]
		s.Delete(seen[0].ID)
		s.Delete(seen[3].ID)
		for _, p := range all {
			if evalLess(last, p) && p.ID%5 == 0 {
				s.Delete(p.ID)
				deleted[p.ID] = true
			}
		}
		s.Add(Product{ID: 1000, Name: "early", CreatedAt: evalBase.Add(-time.Hour)})
		s.Add(Product{ID: 1001, Name: "tie-early", CreatedAt: last.CreatedAt})
		s.Add(Product{ID: 0, Name: "tie-zero", CreatedAt: last.CreatedAt})
	})
	evalCheckOrder(t, seen)
	got := map[int64]bool{}
	for _, p := range seen {
		got[p.ID] = true
	}
	for id := range deleted {
		if got[id] {
			t.Errorf("deleted item %d still returned", id)
		}
	}
	for _, p := range all {
		if !deleted[p.ID] && !got[p.ID] {
			t.Errorf("item %d existed for the whole walk but was skipped", p.ID)
		}
	}
	if got[1000] || got[0] {
		t.Errorf("items inserted behind the cursor were returned")
	}
}

func TestEvalLimitRules(t *testing.T) {
	s, _ := evalSeed(150, 10)
	cases := []struct {
		limit string
		code  int
		n     int
	}{
		{"", 200, 20}, {"0", 200, 20}, {"500", 200, 100}, {"100", 200, 100}, {"35", 200, 35},
		{"-1", 400, 0}, {"abc", 400, 0},
	}
	for _, c := range cases {
		code, items, _, _ := evalPage(t, s, c.limit, "")
		if code != c.code || (code == 200 && len(items) != c.n) {
			t.Errorf("limit %q: status %d with %d items, want %d with %d", c.limit, code, len(items), c.code, c.n)
		}
	}
}

func TestEvalBadCursor(t *testing.T) {
	s, _ := evalSeed(5, 1)
	for _, cursor := range []string{"!!!", "not-a-cursor", "eyJ4IjoxfQ"} {
		if code, _, _, _ := evalPage(t, s, "2", cursor); code != http.StatusBadRequest {
			t.Errorf("cursor %q: status %d, want 400", cursor, code)
		}
	}
}
