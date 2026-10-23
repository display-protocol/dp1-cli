package cmd

import (
	"os"
	"strings"
	"testing"
)

// The version command finds its own version by module path, so a module
// rename that misses cliModulePath makes every installed release print
// "(not in build)". Read go.mod rather than trusting the constant.
func TestCLIModulePathMatchesGoMod(t *testing.T) {
	raw, err := os.ReadFile("../go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if path, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			if got := strings.TrimSpace(path); got != cliModulePath {
				t.Fatalf("go.mod declares module %q, version lookup uses %q", got, cliModulePath)
			}
			return
		}
	}
	t.Fatal("go.mod has no module line")
}
