package fixture_test

import (
	"errors"
	"fixture/api"
	"fixture/users"
	"net/http"
	"net/http/httptest"
	"testing"
)

func evalStore() *users.Store {
	return users.NewStore(users.User{ID: "u1", Name: "Ana", Role: "member"}, users.User{ID: "a1", Name: "Root", Role: "admin"})
}

func TestEvalFind(t *testing.T) {
	s := evalStore()
	if u, err := s.Find("u1"); err != nil || u.Name != "Ana" {
		t.Fatalf("Find(u1) = %+v, %v", u, err)
	}
	if _, err := s.Find("ghost"); !errors.Is(err, users.ErrNotFound) {
		t.Fatalf("Find(ghost) err = %v, want ErrNotFound", err)
	}
}

func TestEvalCallers(t *testing.T) {
	s := evalStore()
	if got := users.Greeting(s, "u1"); got != "Hello, Ana" {
		t.Errorf("Greeting(u1) = %q", got)
	}
	if got := users.Greeting(s, "ghost"); got != "Hello, guest" {
		t.Errorf("Greeting(ghost) = %q, want guest greeting", got)
	}
	cases := []struct {
		id    string
		doc   users.Document
		want  bool
	}{
		{"u1", users.Document{ID: "d", Owner: "u1"}, true},
		{"u1", users.Document{ID: "d", Owner: "x"}, false},
		{"a1", users.Document{ID: "d", Owner: "x"}, true},
		{"ghost", users.Document{ID: "d", Owner: "ghost"}, false},
		{"", users.Document{ID: "d", Owner: ""}, false},
	}
	for _, c := range cases {
		if got := users.CanEdit(s, c.id, c.doc); got != c.want {
			t.Errorf("CanEdit(%q, owner %q) = %v, want %v", c.id, c.doc.Owner, got, c.want)
		}
	}
}

func TestEvalProfile(t *testing.T) {
	h := api.ProfileHandler(evalStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/profile?id=u1", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "Ana" {
		t.Errorf("profile(u1) = %d %q", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/profile?id=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("profile(ghost) status = %d, want 404", rec.Code)
	}
}
