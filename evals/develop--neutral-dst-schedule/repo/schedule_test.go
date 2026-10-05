package schedule

import (
	"testing"
	"time"
)

func TestNextRunUTC(t *testing.T) {
	d := Daily{Hour: 2, Minute: 30}
	after := time.Date(2026, 1, 10, 3, 0, 0, 0, time.UTC)
	want := time.Date(2026, 1, 11, 2, 30, 0, 0, time.UTC)
	if got := d.NextRun(after); !got.Equal(want) {
		t.Fatalf("NextRun = %v, want %v", got, want)
	}
}
