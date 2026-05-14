package create

import (
	"strings"
	"testing"
	"time"
)

func TestSplitComma(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"a", []string{"a"}},
		{"a,b", []string{"a", "b"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"x,,y", []string{"x", "y"}},
	}
	for _, tt := range tests {
		got := splitComma(tt.in)
		if len(got) != len(tt.want) {
			t.Fatalf("splitComma(%q) len got %d want %d", tt.in, len(got), len(tt.want))
		}
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Fatalf("splitComma(%q)[%d] got %q want %q", tt.in, i, got[i], tt.want[i])
			}
		}
	}
}

func TestNowRFC3339(t *testing.T) {
	t.Parallel()
	s := nowRFC3339()
	if !strings.HasSuffix(s, "Z") && !strings.ContainsAny(s, "+-") {
		t.Fatalf("expected UTC-offset RFC3339, got %q", s)
	}
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("Parse RFC3339: %v", err)
	}
	if d := time.Since(ts); d < -time.Minute || d > time.Minute {
		t.Fatalf("timestamp out of range: %v", ts)
	}
}
