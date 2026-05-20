package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/display-protocol/dp1-cli/internal/config"
	"github.com/display-protocol/dp1-cli/internal/output"
)

var initSetupCmd = &cobra.Command{
	Use:   "init",
	Short: `Create ~/.dp1 and a default config file if missing`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := config.Dir(true); err != nil {
			output.PrintError(jsonOut, output.ErrorReport{Command: "init", Error: err.Error()})
			return errPrinted
		}
		p, err := config.Path()
		if err != nil {
			output.PrintError(jsonOut, output.ErrorReport{Command: "init", Error: err.Error()})
			return errPrinted
		}
		_, statErr := os.Stat(p)
		if statErr == nil {
			if jsonOut {
				fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"existed":true,"path":%q}`+"\n", p)
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "config already exists:", p)
			return nil
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			output.PrintError(jsonOut, output.ErrorReport{Command: "init", Error: statErr.Error()})
			return errPrinted
		}

		cfg, loadErr := config.Load()
		if loadErr != nil {
			output.PrintError(jsonOut, output.ErrorReport{Command: "init", Error: loadErr.Error()})
			return errPrinted
		}
		if err := config.Save(cfg); err != nil {
			output.PrintError(jsonOut, output.ErrorReport{Command: "init", Error: err.Error()})
			return errPrinted
		}
		if jsonOut {
			fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"existed":false,"path":%q}`+"\n", p)
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", p)
		fmt.Fprintln(cmd.OutOrStdout(), "next (optional): dp1 key generate --save-config")
		return nil
	},
}
