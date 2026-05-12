package cmd_test

import (
	"io"
	"testing"

	"github.com/spf13/cobra"

	"github.com/display-protocol/dp1-cli/cmd"
)

func childNames(parent *cobra.Command) []string {
	var out []string
	for _, c := range parent.Commands() {
		out = append(out, c.Name())
	}
	return out
}

func mustFindCmd(t *testing.T, root *cobra.Command, path ...string) *cobra.Command {
	t.Helper()
	cur := root
	for _, name := range path {
		var next *cobra.Command
		for _, c := range cur.Commands() {
			if c.Name() == name {
				next = c
				break
			}
		}
		if next == nil {
			t.Fatalf("missing subcommand %q under %q (have: %v)", name, cur.Name(), childNames(cur))
		}
		cur = next
	}
	return cur
}

func assertExecuteFails(t *testing.T, argv []string) {
	t.Helper()
	root := cmd.Root
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	t.Cleanup(func() {
		root.SetArgs(nil)
	})
	root.SetArgs(argv)
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error for args %v", argv)
	}
}
