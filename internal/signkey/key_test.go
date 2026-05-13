package signkey_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/display-protocol/dp1-cli/internal/signkey"
)

func TestParseHexPrivate_seed(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	seed := hex.EncodeToString(priv.Seed())
	got, err := signkey.ParseHexPrivate(seed)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(priv) || !got.Public().(ed25519.PublicKey).Equal(pub) {
		t.Fatal("expanded key mismatch")
	}
}

func TestParseHexPrivate_expandedAnd0xPrefix(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	got, err := signkey.ParseHexPrivate("0x" + hex.EncodeToString(priv))
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(priv) {
		t.Fatal("0x prefixed expanded key mismatch")
	}
}

func TestParseHexPrivate_rejectsBadInput(t *testing.T) {
	for _, s := range []string{"not-hex", "abcd", "aa"} {
		if _, err := signkey.ParseHexPrivate(s); err == nil {
			t.Fatalf("expected error for %q", s)
		}
	}
}
