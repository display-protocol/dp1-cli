package cmd

import (
	"testing"

	"github.com/display-protocol/dp1-cli/internal/config"
)

func TestTrimConfigStdin(t *testing.T) {
	if got := trimConfigStdin([]byte("  abc\nrest")); got != "abc" {
		t.Fatalf("got %q", got)
	}
	if got := trimConfigStdin([]byte("\nkey\n")); got != "key" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyConfigMutation_andPeek(t *testing.T) {
	var cfg config.Config
	if err := applyConfigMutation(&cfg, "signing.private_key", " pk "); err != nil {
		t.Fatal(err)
	}
	v, ok := peekConfig(cfg, "signing.private_key")
	if !ok || v != "pk" {
		t.Fatalf("peek private key: %q ok=%v", v, ok)
	}
	if err := applyConfigMutation(&cfg, "defaults.output_format", "json"); err != nil {
		t.Fatal(err)
	}
	if err := applyConfigMutation(&cfg, "defaults.output_format", "xml"); err == nil {
		t.Fatal("expected bad output_format")
	}
	if err := applyConfigMutation(&cfg, "unknown.key", "x"); err == nil {
		t.Fatal("unknown key")
	}
}
