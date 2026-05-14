package cmd_test

import (
	"strings"
	"testing"

	"github.com/display-protocol/dp1-cli/cmd"
)

func TestChannel_validateRequiresOneSourceArg(t *testing.T) {
	assertExecuteFails(t, []string{"channel", "validate"})
	assertExecuteFails(t, []string{"channel", "validate", "a", "b"})
}

func TestChannel_verifyRequiresOneSourceArg(t *testing.T) {
	assertExecuteFails(t, []string{"channel", "verify"})
	assertExecuteFails(t, []string{"channel", "verify", "a", "b"})
}

func TestChannel_validateAndVerify_useDocumentsSourceArg(t *testing.T) {
	for _, leaf := range []string{"validate", "verify"} {
		c := mustFindCmd(t, cmd.Root, "channel", leaf)
		if !strings.Contains(c.Use, "<source>") {
			t.Fatalf("channel %s: Use line should document <source>: %q", leaf, c.Use)
		}
	}
}

func TestChannel_verify_hasOptionalPubkeyFlag(t *testing.T) {
	c := mustFindCmd(t, cmd.Root, "channel", "verify")
	if c.Args == nil {
		t.Fatal("expected Args validator on channel verify")
	}
	fl := c.LocalFlags().Lookup("pubkey")
	if fl == nil {
		t.Fatal(`expected local --pubkey on channel verify`)
	}
	if fl.DefValue != "" {
		t.Fatalf("pubkey default should be empty, got %q", fl.DefValue)
	}
	if !strings.Contains(fl.Usage, "did:key") {
		t.Fatalf("pubkey usage should mention did:key: %q", fl.Usage)
	}
}

func TestChannel_validate_hasNoPubkeyFlag(t *testing.T) {
	c := mustFindCmd(t, cmd.Root, "channel", "validate")
	if c.LocalFlags().Lookup("pubkey") != nil {
		t.Fatal("channel validate should not define local --pubkey")
	}
}
