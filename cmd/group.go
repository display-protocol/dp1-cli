package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	dp1 "github.com/display-protocol/dp1-go"
	pl "github.com/display-protocol/dp1-go/playlist"
	"github.com/display-protocol/dp1-go/sign"
	"github.com/spf13/cobra"

	"github.com/display-protocol/dp1-cli/internal/create"
	"github.com/display-protocol/dp1-cli/internal/input"
	"github.com/display-protocol/dp1-cli/internal/jsonsign"
	"github.com/display-protocol/dp1-cli/internal/output"
	"github.com/display-protocol/dp1-cli/internal/signkey"
	"github.com/display-protocol/dp1-cli/internal/verify"
)

var groupVerifyPubkey string

var (
	groupCreateOut  string
	groupSignPriv   string
	groupSignRole   string
	groupSignTS     string
	groupSignOutput string
)

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

var groupCreateCmd = &cobra.Command{
	Use:   `create`,
	Short: `Interactively build an unsigned playlist-group JSON draft — run "group sign" before validate`,
	Args:  cobra.NoArgs,
	RunE:  runGroupCreate,
}

var groupSignCmd = &cobra.Command{
	Use:   `sign <file>`,
	Short: `Append one Ed25519 multi-signature to playlist-group JSON`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGroupSign,
}

func init() {
	groupVerifyCmd.Flags().StringVar(&groupVerifyPubkey, "pubkey", "", "Verify only signatures matching this key (did:key, did:pkh, Ed25519 hex, or Ethereum address)")

	groupCreateCmd.Flags().StringVarP(&groupCreateOut, "output", "o", "", "Write draft JSON here (empty or - for stdout)")
	groupSignCmd.Flags().StringVar(&groupSignPriv, "private-key", "", "Hex Ed25519 private key (instead of DP1_PRIVATE_KEY / config)")
	groupSignCmd.Flags().StringVar(&groupSignRole, "role", pl.RoleCurator, "Signature role: curator, feed, agent, institution, licensor")
	groupSignCmd.Flags().StringVar(&groupSignTS, "ts", "", "RFC3339 timestamp (default: now)")
	groupSignCmd.Flags().StringVarP(&groupSignOutput, "output", "o", "", "Write signed JSON here (default: overwrite input; - stdout)")

	groupCmd.AddCommand(groupValidateCmd)
	groupCmd.AddCommand(groupVerifyCmd)
	groupCmd.AddCommand(groupCreateCmd)
	groupCmd.AddCommand(groupSignCmd)
}

func runGroupValidate(cmd *cobra.Command, args []string) error {
	data, err := input.ReadSource(cmd.Context(), args[0])
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
	data, err := input.ReadSource(cmd.Context(), args[0])
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

func runGroupCreate(cmd *cobra.Command, args []string) error {
	g, err := create.Group()
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group create", Error: err.Error()})
		return errPrinted
	}

	dest := strings.TrimSpace(groupCreateOut)
	if err := writeJSONDocument(dest, g, 0o644); err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group create", Error: err.Error()})
		return errPrinted
	}

	if jsonOut {
		fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"draft":true,"resource":"playlist-group","path":%q}`+"\n", draftPathLabel(dest))
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "wrote unsigned draft to %s\n", draftPathLabel(dest))
	fmt.Fprintln(cmd.OutOrStdout(), `Next: dp1 group sign <file>`)
	return nil
}

func runGroupSign(cmd *cobra.Command, args []string) error {
	path := args[0]
	raw, err := os.ReadFile(path)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group sign", Error: err.Error()})
		return errPrinted
	}
	priv, err := signkey.LoadEd25519Private(groupSignPriv)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group sign", Error: err.Error()})
		return errPrinted
	}
	role, err := parseMultiSigRole(groupSignRole)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group sign", Error: err.Error()})
		return errPrinted
	}
	ts := strings.TrimSpace(groupSignTS)
	if ts == "" {
		ts = time.Now().UTC().Format(time.RFC3339)
	}

	signed, err := jsonsign.AppendEd25519(raw, priv, role, ts, jsonsign.ValidatePlaylistGroup)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group sign", Error: err.Error()})
		return errPrinted
	}

	outPath := groupSignOutput
	if outPath == "" {
		outPath = path
	}
	if err := writeRawDocument(cmd, outPath, signed); err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "group sign", Error: err.Error()})
		return errPrinted
	}

	if jsonOut {
		fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"resource":"playlist-group","path":%q}`+"\n", draftPathLabel(outPath))
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "signed playlist-group → %s\n", draftPathLabel(outPath))
	return nil
}
