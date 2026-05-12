package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
)

type ValidateOK struct {
	OK        bool   `json:"ok"`
	Resource  string `json:"resource"`
	DPVersion string `json:"dpVersion,omitempty"` // core playlist / document "dpVersion"
	Version   string `json:"version,omitempty"`   // channels extension "version"
	ID        string `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	Message   string `json:"message,omitempty"`
}

type VerifyOK struct {
	OK          bool   `json:"ok"`
	Resource    string `json:"resource"`
	Mode        string `json:"mode,omitempty"` // "all" | "pubkey" | "legacy"
	Message     string `json:"message,omitempty"`
	PubkeyMatch bool   `json:"pubkeyMatch,omitempty"`
}

type ErrorReport struct {
	OK       bool   `json:"ok"`
	Resource string `json:"resource,omitempty"`
	Command  string `json:"command,omitempty"`
	Error    string `json:"error"`
}

func PrintValidateSuccess(jsonOut bool, v ValidateOK) {
	if jsonOut {
		v.OK = true
		emitJSON(v)
		return
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(os.Stdout, "%s %s is valid\n", green("✓"), v.Resource)
	if v.DPVersion != "" {
		fmt.Fprintf(os.Stdout, "  dpVersion: %s\n", v.DPVersion)
	}
	if v.Version != "" {
		fmt.Fprintf(os.Stdout, "  version:   %s\n", v.Version)
	}
	if v.ID != "" {
		fmt.Fprintf(os.Stdout, "  id:        %s\n", v.ID)
	}
	if v.Title != "" {
		fmt.Fprintf(os.Stdout, "  title:     %s\n", v.Title)
	}
}

func PrintVerifySuccess(jsonOut bool, v VerifyOK) {
	if jsonOut {
		v.OK = true
		emitJSON(v)
		return
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(os.Stdout, "%s Signatures verified (%s)\n", green("✓"), v.Resource)
	if v.Message != "" {
		fmt.Fprintf(os.Stdout, "  %s\n", v.Message)
	}
}

func PrintError(jsonOut bool, rep ErrorReport) {
	if jsonOut {
		rep.OK = false
		emitJSON(rep)
		return
	}
	red := color.New(color.FgRed).SprintFunc()
	if rep.Command != "" {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", red("✗"), rep.Command, rep.Error)
		return
	}
	fmt.Fprintf(os.Stderr, "%s %s\n", red("✗"), rep.Error)
}

func emitJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
