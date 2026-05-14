package cmd

import (
	"github.com/spf13/cobra"
)

var jsonOut bool

// Root is the dp1 root command.
var Root = &cobra.Command{
	Use:           "dp1",
	Short:         "Work with DP-1 playlists, channels, and groups from the command line",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return Root.Execute()
}

func init() {
	Root.PersistentFlags().BoolVar(&jsonOut, "json", false, "Emit JSON on stdout instead of human-readable text")

	Root.AddCommand(versionCmd)
	Root.AddCommand(playlistCmd)
	Root.AddCommand(channelCmd)
	Root.AddCommand(groupCmd)
	Root.AddCommand(configCmd)
	Root.AddCommand(keyCmd)
	Root.AddCommand(initSetupCmd)
}
