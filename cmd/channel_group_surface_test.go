package cmd_test

import (
	"testing"

	"github.com/display-protocol/dp1-cli/cmd"
)

func TestChannel_createSignSurface(t *testing.T) {
	_ = mustFindCmd(t, cmd.Root, "channel", "create")
	s := mustFindCmd(t, cmd.Root, "channel", "sign")
	if s.Flag("output") == nil {
		t.Fatal("expected --output on channel sign")
	}
	fl := s.Flags().Lookup("role")
	if fl == nil {
		t.Fatal("expected --role on channel sign")
	}
	if fl.DefValue != "publisher" {
		t.Fatalf("channel sign --role default: got %q, want publisher", fl.DefValue)
	}
}

func TestGroup_createSignSurface(t *testing.T) {
	_ = mustFindCmd(t, cmd.Root, "group", "create")
	s := mustFindCmd(t, cmd.Root, "group", "sign")
	if s.Flag("role") == nil {
		t.Fatal("expected --role on group sign")
	}
}

func TestChannel_publish_registered(t *testing.T) {
	_ = mustFindCmd(t, cmd.Root, "channel", "publish")
}

func TestGroup_publish_registered(t *testing.T) {
	_ = mustFindCmd(t, cmd.Root, "group", "publish")
}
