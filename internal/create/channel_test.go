package create

import (
	"strings"
	"testing"
)

func TestValidateChannelSummary(t *testing.T) {
	t.Parallel()
	if err := validateChannelSummary(""); err != nil {
		t.Fatal(err)
	}
	if err := validateChannelSummary("  "); err != nil {
		t.Fatal(err)
	}
	if err := validateChannelSummary(strings.Repeat("a", 2000)); err != nil {
		t.Fatal(err)
	}
	if err := validateChannelSummary(strings.Repeat("a", 2001)); err == nil {
		t.Fatal("expected error for too long")
	}
}
