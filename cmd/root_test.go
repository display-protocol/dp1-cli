package cmd_test

import (
	"testing"

	"github.com/display-protocol/dp1-cli/cmd"
)

func TestRoot_cliSurface(t *testing.T) {
	root := cmd.Root
	if root.Name() != "dp1" {
		t.Fatalf("root name: got %q", root.Name())
	}
	names := childNames(root)
	required := []string{"channel", "group", "playlist", "version"}
	for _, want := range required {
		if !stringInSlice(names, want) {
			t.Fatalf("missing top-level command %q, have %v", want, names)
		}
	}

	flag := root.PersistentFlags().Lookup("json")
	if flag == nil {
		t.Fatal(`expected persistent flag --json`)
	}
	if flag.Shorthand != "" {
		t.Fatalf("unexpected --json shorthand %q", flag.Shorthand)
	}
}

func stringInSlice(ss []string, x string) bool {
	for _, s := range ss {
		if s == x {
			return true
		}
	}
	return false
}
