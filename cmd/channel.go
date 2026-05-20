package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	dp1 "github.com/display-protocol/dp1-go"
	dp1ch "github.com/display-protocol/dp1-go/extension/channels"
	"github.com/display-protocol/dp1-go/sign"
	"github.com/spf13/cobra"

	"github.com/display-protocol/dp1-cli/internal/create"
	"github.com/display-protocol/dp1-cli/internal/input"
	"github.com/display-protocol/dp1-cli/internal/jsonsign"
	"github.com/display-protocol/dp1-cli/internal/output"
	"github.com/display-protocol/dp1-cli/internal/signkey"
	"github.com/display-protocol/dp1-cli/internal/verify"
)

var channelVerifyPubkey string

var (
	channelCreateOut  string
	channelSignPriv   string
	channelSignRole   string
	channelSignTS     string
	channelSignOutput string
)

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

var channelCreateCmd = &cobra.Command{
	Use:   `create`,
	Short: `Interactively build an unsigned channel JSON draft — run "channel sign" before validate`,
	Args:  cobra.NoArgs,
	RunE:  runChannelCreate,
}

var channelSignCmd = &cobra.Command{
	Use:   `sign <file>`,
	Short: `Append one Ed25519 multi-signature to channel JSON`,
	Args:  cobra.ExactArgs(1),
	RunE:  runChannelSign,
}

func init() {
	channelVerifyCmd.Flags().StringVar(&channelVerifyPubkey, "pubkey", "", "Verify only signatures matching this key (did:key, did:pkh, Ed25519 hex, or Ethereum address)")

	channelCreateCmd.Flags().StringVarP(&channelCreateOut, "output", "o", "", "Write draft JSON here (empty or - for stdout)")
	channelSignCmd.Flags().StringVar(&channelSignPriv, "private-key", "", "Hex Ed25519 private key (instead of DP1_PRIVATE_KEY / config)")
	channelSignCmd.Flags().StringVar(&channelSignRole, "role", dp1ch.RolePublisher, "Signature role: publisher (default), curator, feed, agent, institution, licensor")
	channelSignCmd.Flags().StringVar(&channelSignTS, "ts", "", "RFC3339 timestamp (default: now)")
	channelSignCmd.Flags().StringVarP(&channelSignOutput, "output", "o", "", "Write signed JSON here (default: overwrite input; - stdout)")

	channelCmd.AddCommand(channelValidateCmd)
	channelCmd.AddCommand(channelVerifyCmd)
	channelCmd.AddCommand(channelCreateCmd)
	channelCmd.AddCommand(channelSignCmd)
}

func runChannelValidate(cmd *cobra.Command, args []string) error {
	data, err := input.ReadSource(cmd.Context(), args[0])
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel validate", Error: err.Error()})
		return errPrinted
	}
	out, err := validateChannel(data, validateAllowUnsigned)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel validate", Error: err.Error()})
		return errPrinted
	}
	output.PrintValidateSuccess(jsonOut, out)
	return nil
}

func runChannelVerify(cmd *cobra.Command, args []string) error {
	data, err := input.ReadSource(cmd.Context(), args[0])
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

func runChannelCreate(cmd *cobra.Command, args []string) error {
	ch, err := create.Channel()
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel create", Error: err.Error()})
		return errPrinted
	}

	dest := strings.TrimSpace(channelCreateOut)
	if err := writeJSONDocument(dest, ch, 0o644); err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel create", Error: err.Error()})
		return errPrinted
	}

	if jsonOut {
		fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"draft":true,"resource":"channel","path":%q}`+"\n", draftPathLabel(dest))
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "wrote unsigned draft to %s\n", draftPathLabel(dest))
	fmt.Fprintln(cmd.OutOrStdout(), `Next: dp1 channel sign <file>`)
	return nil
}

func runChannelSign(cmd *cobra.Command, args []string) error {
	path := args[0]
	raw, err := os.ReadFile(path)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel sign", Error: err.Error()})
		return errPrinted
	}
	priv, err := signkey.LoadEd25519Private(channelSignPriv)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel sign", Error: err.Error()})
		return errPrinted
	}
	role, err := parseMultiSigRole(channelSignRole)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel sign", Error: err.Error()})
		return errPrinted
	}
	ts := strings.TrimSpace(channelSignTS)
	if ts == "" {
		ts = time.Now().UTC().Format(time.RFC3339)
	}

	signed, err := jsonsign.AppendEd25519(raw, priv, role, ts, jsonsign.ValidateChannel)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel sign", Error: err.Error()})
		return errPrinted
	}

	outPath := channelSignOutput
	if outPath == "" {
		outPath = path
	}
	if err := writeRawDocument(cmd, outPath, signed); err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "channel sign", Error: err.Error()})
		return errPrinted
	}

	if jsonOut {
		fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"resource":"channel","path":%q}`+"\n", draftPathLabel(outPath))
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "signed channel → %s\n", draftPathLabel(outPath))
	return nil
}
