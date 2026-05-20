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
	required := []string{"channel", "config", "group", "init", "key", "playlist", "version"}
	for _, want := range required {
		if !stringInSlice(names, want) {
			t.Fatalf("missing top-level command %q, have %v", want, names)
		}
	}
	seen := make(map[string]int)
	for _, n := range names {
		seen[n]++
	}
	for name, count := range seen {
		if count != 1 {
			t.Fatalf("top-level command %q registered %d times, have %v", name, count, names)
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
