package cmd

import (
	"errors"
	"fmt"

	dp1 "github.com/display-protocol/dp1-go"
	"github.com/display-protocol/dp1-go/sign"
	"github.com/spf13/cobra"

	"github.com/display-protocol/dp1-cli/internal/input"
	"github.com/display-protocol/dp1-cli/internal/output"
	"github.com/display-protocol/dp1-cli/internal/verify"
)

var groupVerifyPubkey string

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "DP-1 playlist-group (exhibition) commands",
}

var groupValidateCmd = &cobra.Command{
	Use:   "validate <source>",
	Short: "Validate playlist-group JSON against the DP-1 schema (from file, URL, stdin, or base64)",
	Args:  cobra.ExactArgs(1),
	RunE:  runGroupValidate,
}

var groupVerifyCmd = &cobra.Command{
	Use:   "verify <source>",
	Short: "Validate playlist-group JSON and verify cryptographic signatures",
	Args:  cobra.ExactArgs(1),
	RunE:  runGroupVerify,
}

func init() {
	groupVerifyCmd.Flags().StringVar(&groupVerifyPubkey, "pubkey", "", "Verify only signatures matching this key (did:key, did:pkh, Ed25519 hex, or Ethereum address)")

	groupCmd.AddCommand(groupValidateCmd)
	groupCmd.AddCommand(groupVerifyCmd)
}

func runGroupValidate(cmd *cobra.Command, args []string) error {
	data, err := input.ReadSource(args[0])
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group validate", Error: err.Error()})
		return errPrinted
	}
	g, err := dp1.ParseAndValidatePlaylistGroup(data)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group validate", Error: err.Error()})
		return errPrinted
	}
	output.PrintValidateSuccess(jsonOut, output.ValidateOK{
		Resource:  "playlist-group",
		DPVersion: "",
		ID:        g.ID,
		Title:     g.Title,
	})
	return nil
}

func runGroupVerify(cmd *cobra.Command, args []string) error {
	data, err := input.ReadSource(args[0])
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group verify", Error: err.Error()})
		return errPrinted
	}
	if _, err := dp1.ParseAndValidatePlaylistGroup(data); err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group verify", Error: err.Error()})
		return errPrinted
	}
	err = verify.Run(data, verify.PlaylistGroup, groupVerifyPubkey)
	if err != nil {
		if errors.Is(err, sign.ErrNoSignatures) {
			err = fmt.Errorf("%w: add v1.1+ \"signatures\" or legacy \"signature\" with --pubkey for legacy", err)
		}
		output.PrintError(jsonOut, output.ErrorReport{Command: "group verify", Error: err.Error()})
		return errPrinted
	}
	msg := "all signatures in the document"
	if groupVerifyPubkey != "" {
		msg = "signatures matching --pubkey"
	}
	output.PrintVerifySuccess(jsonOut, output.VerifyOK{
		Resource: "playlist-group",
		Mode:     modeLabel(groupVerifyPubkey != ""),
		Message:  msg,
	})
	return nil
}
