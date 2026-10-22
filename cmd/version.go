package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// cliModulePath is this module's own path, as go.mod declares it. The
// version lookup matches the build info's main module against it, so it must
// carry the /v2 major-version suffix: without it every installed release
// reports "(not in build)". TestCLIModulePathMatchesGoMod keeps the two in step.
const cliModulePath = "github.com/display-protocol/dp1-cli/v2"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print dp1-cli and dependency versions",
	RunE:  runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	cli := moduleVersion(cliModulePath)
	lib := moduleVersion("github.com/display-protocol/dp1-go")
	out := cmd.OutOrStdout()
	if jsonOut {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]string{
			"dp1_cli": cli,
			"dp1_go":  lib,
			"go":      runtime.Version(),
		})
	}
	fmt.Fprintf(out, "dp1-cli %s\n", cli)
	fmt.Fprintf(out, "dp1-go library %s\n", lib)
	fmt.Fprintf(out, "%s\n", runtime.Version())
	return nil
}

func moduleVersion(modulePath string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unknown)"
	}
	if info.Main.Path == modulePath {
		v := strings.TrimPrefix(info.Main.Version, "(devel)")
		if strings.TrimSpace(v) == "" {
			return "(devel)"
		}
		return strings.TrimSpace(v)
	}
	for _, dep := range info.Deps {
		if dep.Path == modulePath {
			if dep.Replace != nil {
				return dep.Version + " => " + dep.Replace.Path
			}
			return dep.Version
		}
	}
	return "(not in build)"
}
