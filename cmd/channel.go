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

var channelVerifyPubkey string

var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "DP-1 channels extension commands",
}

var channelValidateCmd = &cobra.Command{
	Use:   "validate <source>",
	Short: "Validate channel JSON against the DP-1 channels schema (from file, URL, stdin, or base64)",
	Args:  cobra.ExactArgs(1),
	RunE:  runChannelValidate,
}

var channelVerifyCmd = &cobra.Command{
	Use:   "verify <source>",
	Short: "Validate channel JSON and verify cryptographic signatures",
	Args:  cobra.ExactArgs(1),
	RunE:  runChannelVerify,
}

func init() {
	channelVerifyCmd.Flags().StringVar(&channelVerifyPubkey, "pubkey", "", "Verify only signatures matching this key (did:key, did:pkh, Ed25519 hex, or Ethereum address)")

	channelCmd.AddCommand(channelValidateCmd)
	channelCmd.AddCommand(channelVerifyCmd)
}

func runChannelValidate(cmd *cobra.Command, args []string) error {
	data, err := input.ReadSource(args[0])
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel validate", Error: err.Error()})
		return errPrinted
	}
	ch, err := dp1.ParseAndValidateChannel(data)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel validate", Error: err.Error()})
		return errPrinted
	}
	output.PrintValidateSuccess(jsonOut, output.ValidateOK{
		Resource: "channel",
		Version:  ch.Version,
		ID:       ch.ID,
		Title:    ch.Title,
	})
	return nil
}

func runChannelVerify(cmd *cobra.Command, args []string) error {
	data, err := input.ReadSource(args[0])
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel verify", Error: err.Error()})
		return errPrinted
	}
	if _, err := dp1.ParseAndValidateChannel(data); err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel verify", Error: err.Error()})
		return errPrinted
	}
	err = verify.Run(data, verify.Channel, channelVerifyPubkey)
	if err != nil {
		if errors.Is(err, sign.ErrNoSignatures) {
			err = fmt.Errorf("%w: add v1.1+ \"signatures\" or legacy \"signature\" with --pubkey for legacy", err)
		}
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel verify", Error: err.Error()})
		return errPrinted
	}
	msg := "all signatures in the document"
	if channelVerifyPubkey != "" {
		msg = "signatures matching --pubkey"
	}
	output.PrintVerifySuccess(jsonOut, output.VerifyOK{
		Resource: "channel",
		Mode:     modeLabel(channelVerifyPubkey != ""),
		Message:  msg,
	})
	return nil
}
