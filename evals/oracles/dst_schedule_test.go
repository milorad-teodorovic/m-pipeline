package schedule

import (
	"testing"
	"time"
	_ "time/tzdata"
)

func evalZone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func evalUTC(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestEvalNextRunInZone(t *testing.T) {
	ny := evalZone(t, "America/New_York")
	berlin := evalZone(t, "Europe/Berlin")
	tokyo := evalZone(t, "Asia/Tokyo")
	cases := []struct {
		name  string
		job   Daily
		after string
		want  string
	}{
		{"nil location is UTC", Daily{Hour: 2, Minute: 30}, "2026-01-10T03:00:00Z", "2026-01-11T02:30:00Z"},
		{"new york winter", Daily{Hour: 2, Minute: 30, Location: ny}, "2026-01-10T12:00:00Z", "2026-01-11T07:30:00Z"},
		{"new york skipped time runs at the change", Daily{Hour: 2, Minute: 30, Location: ny}, "2026-03-07T12:00:00Z", "2026-03-08T07:00:00Z"},
		{"new york day after the change", Daily{Hour: 2, Minute: 30, Location: ny}, "2026-03-08T07:00:00Z", "2026-03-09T06:30:00Z"},
		{"new york repeated time first occurrence", Daily{Hour: 1, Minute: 30, Location: ny}, "2026-11-01T04:00:00Z", "2026-11-01T05:30:00Z"},
		{"new york repeated time runs once", Daily{Hour: 1, Minute: 30, Location: ny}, "2026-11-01T05:30:00Z", "2026-11-02T06:30:00Z"},
		{"berlin skipped time runs at the change", Daily{Hour: 2, Minute: 30, Location: berlin}, "2026-03-28T12:00:00Z", "2026-03-29T01:00:00Z"},
		{"berlin repeated time first occurrence", Daily{Hour: 2, Minute: 30, Location: berlin}, "2026-10-24T12:00:00Z", "2026-10-25T00:30:00Z"},
		{"berlin repeated time runs once", Daily{Hour: 2, Minute: 30, Location: berlin}, "2026-10-25T00:30:00Z", "2026-10-26T01:30:00Z"},
		{"tokyo local date differs from utc date", Daily{Hour: 2, Minute: 30, Location: tokyo}, "2026-01-10T20:00:00Z", "2026-01-11T17:30:00Z"},
		{"tokyo just before the run", Daily{Hour: 2, Minute: 30, Location: tokyo}, "2026-01-11T17:29:00Z", "2026-01-11T17:30:00Z"},
	}
	for _, c := range cases {
		got := c.job.NextRun(evalUTC(c.after))
		if want := evalUTC(c.want); !got.Equal(want) {
			t.Errorf("%s: NextRun(%s) = %s, want %s", c.name, c.after, got.UTC().Format(time.RFC3339), c.want)
		}
	}
}

func TestEvalOneRunPerDay(t *testing.T) {
	for _, name := range []string{"America/New_York", "Europe/Berlin", "Australia/Lord_Howe"} {
		loc := evalZone(t, name)
		for _, job := range []Daily{{1, 30, loc}, {2, 0, loc}, {2, 30, loc}} {
			at := evalUTC("2026-01-01T00:00:00Z")
			var prev time.Time
			for i := 0; i < 400; i++ {
				next := job.NextRun(at)
				if !next.After(at) {
					t.Fatalf("%s %02d:%02d: NextRun(%s) = %s is not after", name, job.Hour, job.Minute, at, next)
				}
				if !prev.IsZero() {
					py, pm, pd := prev.In(loc).Date()
					ny, nm, nd := next.In(loc).Date()
					gap := time.Date(ny, nm, nd, 0, 0, 0, 0, time.UTC).Sub(time.Date(py, pm, pd, 0, 0, 0, 0, time.UTC))
					if gap != 24*time.Hour {
						t.Fatalf("%s %02d:%02d: runs %s and %s are not on consecutive days", name, job.Hour, job.Minute, prev.In(loc), next.In(loc))
					}
				}
				prev, at = next, next
			}
		}
	}
}
