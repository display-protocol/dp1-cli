package cmd_test

import (
	"strings"
	"testing"

	"github.com/display-protocol/dp1-cli/cmd"
)

func TestGroup_validateRequiresOneSourceArg(t *testing.T) {
	assertExecuteFails(t, []string{"group", "validate"})
	assertExecuteFails(t, []string{"group", "validate", "a", "b"})
}

func TestGroup_verifyRequiresOneSourceArg(t *testing.T) {
	assertExecuteFails(t, []string{"group", "verify"})
	assertExecuteFails(t, []string{"group", "verify", "a", "b"})
}

func TestGroup_validateAndVerify_useDocumentsSourceArg(t *testing.T) {
	for _, leaf := range []string{"validate", "verify"} {
		c := mustFindCmd(t, cmd.Root, "group", leaf)
		if !strings.Contains(c.Use, "<source>") {
			t.Fatalf("group %s: Use line should document <source>: %q", leaf, c.Use)
		}
	}
}

func TestGroup_verify_hasOptionalPubkeyFlag(t *testing.T) {
	c := mustFindCmd(t, cmd.Root, "group", "verify")
	if c.Args == nil {
		t.Fatal("expected Args validator on group verify")
	}
	fl := c.LocalFlags().Lookup("pubkey")
	if fl == nil {
		t.Fatal(`expected local --pubkey on group verify`)
	}
	if fl.DefValue != "" {
		t.Fatalf("pubkey default should be empty, got %q", fl.DefValue)
	}
	if !strings.Contains(fl.Usage, "did:key") {
		t.Fatalf("pubkey usage should mention did:key: %q", fl.Usage)
	}
}

func TestGroup_validate_hasNoPubkeyFlag(t *testing.T) {
	c := mustFindCmd(t, cmd.Root, "group", "validate")
	if c.LocalFlags().Lookup("pubkey") != nil {
		t.Fatal("group validate should not define local --pubkey")
	}
}

func TestGroup_createSignSurface(t *testing.T) {
	_ = mustFindCmd(t, cmd.Root, "group", "create")
	s := mustFindCmd(t, cmd.Root, "group", "sign")
	if s.Flag("output") == nil {
		t.Fatal("expected --output on group sign")
	}
	fl := s.Flags().Lookup("role")
	if fl == nil {
		t.Fatal("expected --role on group sign")
	}
	if fl.DefValue != "curator" {
		t.Fatalf("group sign --role default: got %q, want curator", fl.DefValue)
	}
}

func TestGroup_publish_registered(t *testing.T) {
	_ = mustFindCmd(t, cmd.Root, "group", "publish")
}
