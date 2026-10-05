package catalog

import (
	"reflect"
	"testing"
)

func TestEvalPage(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	tests := []struct {
		page, size int
		want       []string
	}{
		{1, 2, []string{"a", "b"}},
		{2, 2, []string{"c", "d"}},
		{3, 2, []string{"e"}},
		{4, 2, []string{}},
		{100, 2, []string{}},
		{1, 5, []string{"a", "b", "c", "d", "e"}},
		{1, 10, []string{"a", "b", "c", "d", "e"}},
		{0, 2, []string{}},
		{-1, 2, []string{}},
		{1, 0, []string{}},
		{1, -3, []string{}},
	}
	for _, tt := range tests {
		got := Page(items, tt.page, tt.size)
		if len(got) == 0 && len(tt.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Page(%d, %d) = %v, want %v", tt.page, tt.size, got, tt.want)
		}
	}
	if got := Page(nil, 1, 2); len(got) != 0 {
		t.Errorf("Page(nil) = %v, want empty", got)
	}
}
