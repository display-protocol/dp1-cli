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

var playlistVerifyPubkey string

var (
	playlistCreateOut  string
	playlistSignPriv   string
	playlistSignRole   string
	playlistSignTS     string
	playlistSignOutput string
)

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

var playlistCreateCmd = &cobra.Command{
	Use:   `create`,
	Short: `Interactively build an unsigned playlist JSON draft (schemas require signatures — run "playlist sign" next)`,
	Args:  cobra.NoArgs,
	RunE:  runPlaylistCreate,
}

var playlistSignCmd = &cobra.Command{
	Use:   `sign <file>`,
	Short: "Append one Ed25519 multi-signature to a playlist JSON document (preserves unknown fields)",
	Args:  cobra.ExactArgs(1),
	RunE:  runPlaylistSign,
}

func init() {
	playlistVerifyCmd.Flags().StringVar(&playlistVerifyPubkey, "pubkey", "", "Verify only signatures matching this key (did:key, did:pkh, Ed25519 hex, or Ethereum address)")

	playlistCreateCmd.Flags().StringVarP(&playlistCreateOut, "output", "o", "", "Write draft JSON here (empty or - for stdout)")
	playlistSignCmd.Flags().StringVar(&playlistSignPriv, "private-key", "", "Hex Ed25519 private key (instead of DP1_PRIVATE_KEY / config)")
	playlistSignCmd.Flags().StringVar(&playlistSignRole, "role", pl.RoleCurator, "Signature role: curator, feed, agent, institution, licensor")
	playlistSignCmd.Flags().StringVar(&playlistSignTS, "ts", "", "RFC3339 timestamp (UTC recommended; default: now)")
	playlistSignCmd.Flags().StringVarP(&playlistSignOutput, "output", "o", "", "Write signed JSON here (default: overwrite <file>; - for stdout)")

	playlistCmd.AddCommand(playlistValidateCmd)
	playlistCmd.AddCommand(playlistVerifyCmd)
	playlistCmd.AddCommand(playlistCreateCmd)
	playlistCmd.AddCommand(playlistSignCmd)
}

func runPlaylistValidate(cmd *cobra.Command, args []string) error {
	data, err := input.ReadSource(cmd.Context(), args[0])
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
	data, err := input.ReadSource(cmd.Context(), args[0])
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

func runPlaylistCreate(cmd *cobra.Command, args []string) error {
	p, err := create.Playlist()
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist create", Error: err.Error()})
		return errPrinted
	}

	dest := strings.TrimSpace(playlistCreateOut)
	if err := writeJSONDocument(dest, p, 0o644); err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist create", Error: err.Error()})
		return errPrinted
	}

	if jsonOut {
		fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"draft":true,"resource":"playlist","path":%q}`+"\n", draftPathLabel(dest))
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "wrote unsigned draft to %s\n", draftPathLabel(dest))
	fmt.Fprintln(cmd.OutOrStdout(), `Next: dp1 playlist sign <file>`)
	return nil
}

func runPlaylistSign(cmd *cobra.Command, args []string) error {
	path := args[0]
	raw, err := os.ReadFile(path)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist sign", Error: err.Error()})
		return errPrinted
	}
	priv, err := signkey.LoadEd25519Private(playlistSignPriv)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist sign", Error: err.Error()})
		return errPrinted
	}
	role, err := parseMultiSigRole(playlistSignRole)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist sign", Error: err.Error()})
		return errPrinted
	}
	ts := strings.TrimSpace(playlistSignTS)
	if ts == "" {
		ts = time.Now().UTC().Format(time.RFC3339)
	}

	signed, err := jsonsign.AppendEd25519(raw, priv, role, ts, jsonsign.ValidatePlaylist)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist sign", Error: err.Error()})
		return errPrinted
	}

	outPath := playlistSignOutput
	if outPath == "" {
		outPath = path
	}
	if err := writeRawDocument(cmd, outPath, signed); err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "playlist sign", Error: err.Error()})
		return errPrinted
	}

	if jsonOut {
		fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"resource":"playlist","path":%q}`+"\n", draftPathLabel(outPath))
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "signed playlist → %s\n", draftPathLabel(outPath))
	return nil
}

func draftPathLabel(dest string) string {
	if dest == "" || dest == "-" {
		return "-"
	}
	return dest
}

func modeLabel(pubkey bool) string {
	if pubkey {
		return "pubkey"
	}
	return "all"
}
