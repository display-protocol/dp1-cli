package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/display-protocol/dp1-go/sign"
	"github.com/spf13/cobra"

	"github.com/display-protocol/dp1-cli/internal/ask"
	"github.com/display-protocol/dp1-cli/internal/config"
	"github.com/display-protocol/dp1-cli/internal/output"
	"github.com/display-protocol/dp1-cli/internal/signkey"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Generate and inspect Ed25519 keys used by dp1-cli",
}

var (
	keyGenerateSave bool
	keyImportHex    string
	keyShowPriv     string
)

var keyGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an Ed25519 key pair suitable for playlist-style multi-signatures",
	Args:  cobra.NoArgs,
	RunE:  runKeyGenerate,
}

var keyImportCmd = &cobra.Command{
	Use:   `import [HEX_PRIVATE_KEY]`,
	Short: `Store a private key in ~/.dp1/config.yaml (64-hex seed or 128-hex expanded); omit HEX to paste interactively`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runKeyImport,
}

var keyShowCmd = &cobra.Command{
	Use:   "show",
	Short: `Show public key derived from --private-key, DP1_PRIVATE_KEY, or config (never prints secrets)`,
	Args:  cobra.NoArgs,
	RunE:  runKeyShow,
}

func init() {
	keyGenerateCmd.Flags().BoolVar(&keyGenerateSave, "save-config", false, "Write signing.private_key to ~/.dp1/config.yaml")
	keyImportCmd.Flags().StringVar(&keyImportHex, "private-key", "", "Explicit hex instead of HEX arg / interactive paste")
	keyShowCmd.Flags().StringVar(&keyShowPriv, "private-key", "", "Inspect this hex key instead of config/env defaults")

	keyCmd.AddCommand(keyGenerateCmd)
	keyCmd.AddCommand(keyImportCmd)
	keyCmd.AddCommand(keyShowCmd)
	Root.AddCommand(keyCmd)
}

func runKeyGenerate(cmd *cobra.Command, args []string) error {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "key generate", Error: err.Error()})
		return errPrinted
	}

	did, err := sign.Ed25519DIDKey(pub)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "key generate", Error: err.Error()})
		return errPrinted
	}

	hexExpanded := hex.EncodeToString(priv)
	if jsonOut {
		fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"public_did_key":%q,"public_hex":%q,"private_key_hex_expanded":%q}`+"\n",
			did, hex.EncodeToString(pub), hexExpanded)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "public kid (did:key): %s\n", did)
		fmt.Fprintf(cmd.OutOrStdout(), "public hex: %s\n", hex.EncodeToString(pub))
		fmt.Fprintf(cmd.OutOrStdout(), "private key hex (128 chars, KEEP SECRET): %s\n", hexExpanded)
	}

	if keyGenerateSave {
		cfg, err := config.Load()
		if err != nil {
			output.PrintError(jsonOut, output.ErrorReport{Command: "key generate", Error: err.Error()})
			return errPrinted
		}
		cfg.Signing.PrivateKey = hexExpanded
		cfg.Signing.PublicKey = hex.EncodeToString(pub)
		if err := config.Save(cfg); err != nil {
			output.PrintError(jsonOut, output.ErrorReport{Command: "key generate", Error: err.Error()})
			return errPrinted
		}
		if !jsonOut {
			fmt.Fprintln(cmd.OutOrStdout(), "saved signing.private_key to ~/.dp1/config.yaml")
		}
	}
	return nil
}

func runKeyImport(cmd *cobra.Command, args []string) error {
	raw := normalizeHexArg(args, keyImportHex)
	if raw == "" {
		var err error
		raw, err = ask.Line("Ed25519 private key hex", "", false, func(s string) error {
			if _, err := signkey.ParseHexPrivate(s); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			output.PrintError(jsonOut, output.ErrorReport{Command: "key import", Error: err.Error()})
			return errPrinted
		}
		raw = normalizeHexArg(nil, raw)
	}
	priv, err := signkey.ParseHexPrivate(raw)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "key import", Error: err.Error()})
		return errPrinted
	}
	pub := priv.Public().(ed25519.PublicKey)
	did, err := sign.Ed25519DIDKey(pub)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "key import", Error: err.Error()})
		return errPrinted
	}

	cfg, err := config.Load()
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "key import", Error: err.Error()})
		return errPrinted
	}
	cfg.Signing.PrivateKey = hex.EncodeToString(priv)
	cfg.Signing.PublicKey = hex.EncodeToString(pub)
	if err := config.Save(cfg); err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "key import", Error: err.Error()})
		return errPrinted
	}

	if jsonOut {
		fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"public_did_key":%q}`+"\n", did)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "stored private key — public kid: %s\n", did)
	return nil
}

func runKeyShow(cmd *cobra.Command, args []string) error {
	var priv ed25519.PrivateKey
	var err error
	switch {
	case normalizeHexArg(nil, keyShowPriv) != "":
		priv, err = signkey.ParseHexPrivate(keyShowPriv)
	default:
		priv, err = signkey.LoadEd25519Private("")
	}
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "key show", Error: err.Error()})
		return errPrinted
	}
	pub := priv.Public().(ed25519.PublicKey)
	did, err := sign.Ed25519DIDKey(pub)
	if err != nil {
		output.PrintError(jsonOut, output.ErrorReport{Command: "key show", Error: err.Error()})
		return errPrinted
	}

	if jsonOut {
		fmt.Fprintf(cmd.OutOrStdout(), `{"ok":true,"public_did_key":%q,"public_hex":%q}`+"\n",
			did, hex.EncodeToString(pub))
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n%s\n", did, hex.EncodeToString(pub))
	return nil
}

func normalizeHexArg(args []string, flag string) string {
	s := strings.TrimSpace(flag)
	if s != "" {
		return strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	}
	if len(args) >= 1 {
		a := strings.TrimSpace(args[0])
		return strings.TrimPrefix(strings.TrimPrefix(a, "0x"), "0X")
	}
	return ""
}
