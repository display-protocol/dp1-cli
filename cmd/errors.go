package cmd

import "errors"

// errPrinted is returned after the CLI prints an error so Cobra exits non-zero without duplicate stderr (Root.SilenceErrors).
var errPrinted = errors.New("")
