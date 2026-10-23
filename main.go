package main

import (
	"os"

	"github.com/display-protocol/dp1-cli/v2/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
