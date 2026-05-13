package signkey

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/display-protocol/dp1-cli/internal/config"

	"crypto/ed25519"
)

const envPrivateKey = "DP1_PRIVATE_KEY"

// LoadEd25519Private resolves an Ed25519 private key:
// precedence: hex flag (--private-key) > DP1_PRIVATE_KEY > ~/.dp1 signing.private_key
func LoadEd25519Private(flagHex string) (ed25519.PrivateKey, error) {
	sources := []string{
		strings.TrimSpace(flagHex),
		strings.TrimSpace(os.Getenv(envPrivateKey)),
	}
	for _, raw := range sources {
		if raw != "" {
			return ParseHexPrivate(raw)
		}
	}
	cfg, err := config.LoadCached()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if pk := strings.TrimSpace(cfg.Signing.PrivateKey); pk != "" {
		return ParseHexPrivate(pk)
	}
	return nil, fmt.Errorf("no signing key: set --private-key, %s, or run `dp1 key import`", envPrivateKey)
}

// ParseHexPrivate accepts a 32-byte seed (64 hex chars) or a 64-byte expanded Ed25519 key (128 hex chars).
func ParseHexPrivate(s string) (ed25519.PrivateKey, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "0x")
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("private key must be hex: %w", err)
	}
	switch len(b) {
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(b), nil
	case ed25519.PrivateKeySize:
		k := ed25519.PrivateKey(append(ed25519.PrivateKey(nil), b...))
		if err := sanityCheckExpanded(k); err != nil {
			return nil, err
		}
		return k, nil
	default:
		return nil, fmt.Errorf("Ed25519 private key must be %d (seed) or %d bytes in hex", ed25519.SeedSize, ed25519.PrivateKeySize)
	}
}

func sanityCheckExpanded(k ed25519.PrivateKey) error {
	pub := k.Public().(ed25519.PublicKey)
	if len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid expanded private key")
	}
	fromSeed := ed25519.NewKeyFromSeed(k.Seed())
	if !fromSeed.Equal(k) {
		return fmt.Errorf("expanded private key does not match embedded seed — use 64-hex-char seed instead")
	}
	return nil
}
