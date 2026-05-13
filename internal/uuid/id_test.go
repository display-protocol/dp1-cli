package uuid

import (
	"strings"
	"testing"

	"github.com/display-protocol/dp1-cli/internal/fields"
)

func TestNewV4_formatAndVariance(t *testing.T) {
	seen := make(map[string]struct{})
	for range 24 {
		u, err := NewV4()
		if err != nil {
			t.Fatal(err)
		}
		if err := fields.UUID(u); err != nil {
			t.Fatalf("invalid uuid %q: %v", u, err)
		}
		if strings.ToLower(u) != u {
			t.Fatalf("expected lowercase: %q", u)
		}
		seen[u] = struct{}{}
	}
	if len(seen) < 20 {
		t.Fatalf("expected mostly unique IDs, got %d distinct", len(seen))
	}
}
