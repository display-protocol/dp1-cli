package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print dp1-cli and dependency versions",
	RunE:  runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	cli := moduleVersion("github.com/display-protocol/dp1-cli")
	lib := moduleVersion("github.com/display-protocol/dp1-go")
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]string{
			"dp1_cli": cli,
			"dp1_go":  lib,
			"go":      runtime.Version(),
		})
	}
	fmt.Printf("dp1-cli %s\n", cli)
	fmt.Printf("dp1-go library %s\n", lib)
	fmt.Printf("%s\n", runtime.Version())
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
