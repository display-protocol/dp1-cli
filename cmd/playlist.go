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

var playlistVerifyPubkey string

var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "DP-1 core playlist commands",
}

var playlistValidateCmd = &cobra.Command{
	Use:   "validate <source>",
	Short: "Validate playlist JSON against the DP-1 core schema (from file, URL, stdin, or base64)",
	Args:  cobra.ExactArgs(1),
	RunE:  runPlaylistValidate,
}

var playlistVerifyCmd = &cobra.Command{
	Use:   "verify <source>",
	Short: "Validate playlist JSON and verify cryptographic signatures (v1.1+ multi-sig and optional v1.0 legacy)",
	Args:  cobra.ExactArgs(1),
	RunE:  runPlaylistVerify,
}

func init() {
	playlistVerifyCmd.Flags().StringVar(&playlistVerifyPubkey, "pubkey", "", "Verify only signatures matching this key (did:key, did:pkh, Ed25519 hex, or Ethereum address)")

	playlistCmd.AddCommand(playlistValidateCmd)
	playlistCmd.AddCommand(playlistVerifyCmd)
}

func runPlaylistValidate(cmd *cobra.Command, args []string) error {
	data, err := input.ReadSource(args[0])
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist validate", Error: err.Error()})
		return errPrinted
	}
	p, err := dp1.ParseAndValidatePlaylist(data)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist validate", Error: err.Error()})
		return errPrinted
	}
	out := output.ValidateOK{
		Resource:  "playlist",
		DPVersion: p.DPVersion,
		ID:        p.ID,
		Title:     p.Title,
	}
	output.PrintValidateSuccess(jsonOut, out)
	return nil
}

func runPlaylistVerify(cmd *cobra.Command, args []string) error {
	data, err := input.ReadSource(args[0])
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist verify", Error: err.Error()})
		return errPrinted
	}
	if _, err := dp1.ParseAndValidatePlaylist(data); err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist verify", Error: err.Error()})
		return errPrinted
	}
	err = verify.Run(data, verify.Playlist, playlistVerifyPubkey)
	if err != nil {
		if errors.Is(err, sign.ErrNoSignatures) {
			err = fmt.Errorf("%w: add v1.1+ \"signatures\" or legacy \"signature\" with --pubkey for legacy", err)
		}
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist verify", Error: err.Error()})
		return errPrinted
	}
	msg := "all signatures in the document"
	if playlistVerifyPubkey != "" {
		msg = "signatures matching --pubkey"
	}
	output.PrintVerifySuccess(jsonOut, output.VerifyOK{
		Resource: "playlist",
		Mode:     modeLabel(playlistVerifyPubkey != ""),
		Message:  msg,
	})
	return nil
}

func modeLabel(pubkey bool) string {
	if pubkey {
		return "pubkey"
	}
	return "all"
}
