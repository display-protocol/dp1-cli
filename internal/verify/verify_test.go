package verify

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/display-protocol/dp1-go/playlist"
	"github.com/display-protocol/dp1-go/sign"
)

func TestParsePubkeyHint(t *testing.T) {
	t.Parallel()
	pub, _, _ := ed25519.GenerateKey(nil)
	hex64 := hex.EncodeToString(pub)
	kid, err := sign.Ed25519DIDKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	pkh := "did:pkh:eip155:1:0xb9c5714089478a327f09197987f16f9e5d936e8a"

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "did:key", in: kid},
		{name: "did:pkh", in: pkh},
		{name: "ed25519_hex", in: hex64},
		{name: "eth_0x", in: "0xb9c5714089478a327f09197987f16f9e5d936e8a"},
		{name: "eth_no_prefix", in: "b9c5714089478a327f09197987f16f9e5d936e8a"},
		{name: "empty", in: "", wantErr: true},
		{name: "garbage", in: "hello", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, err := ParsePubkeyHint(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if h == nil {
				t.Fatal("nil hint")
			}
		})
	}
}

func TestRun_playlist_multiSig_ok(t *testing.T) {
	t.Parallel()
	_, priv, _ := ed25519.GenerateKey(nil)
	pl := playlist.Playlist{
		DPVersion: "1.1.0",
		Title:     "ok",
		Items:     []playlist.PlaylistItem{{Source: "https://example.com"}},
	}
	raw, err := json.Marshal(pl)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := sign.SignMultiEd25519(raw, priv, playlist.RoleCurator, "2025-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	pl.Signatures = []playlist.Signature{sig}
	signed, err := json.Marshal(pl)
	if err != nil {
		t.Fatal(err)
	}
	if err := Run(signed, Playlist, ""); err != nil {
		t.Fatal(err)
	}
}

func TestRun_noSignatures(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"dpVersion":"1.1.0","title":"x","items":[{"source":"https://a"}]}`)
	err := Run(raw, Playlist, "")
	if !errors.Is(err, sign.ErrNoSignatures) {
		t.Fatalf("want ErrNoSignatures, got %v", err)
	}
}

func TestRun_invalidJSON(t *testing.T) {
	t.Parallel()
	err := Run([]byte(`{`), Playlist, "")
	if err == nil || !strings.Contains(err.Error(), "decode document envelope") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRun_pubkeyNoMatch(t *testing.T) {
	t.Parallel()
	_, priv, _ := ed25519.GenerateKey(nil)
	pl := playlist.Playlist{
		DPVersion: "1.1.0",
		Title:     "nm",
		Items:     []playlist.PlaylistItem{{Source: "https://x"}},
	}
	raw, _ := json.Marshal(pl)
	sig, _ := sign.SignMultiEd25519(raw, priv, playlist.RoleCurator, "2025-01-01T00:00:00Z")
	pl.Signatures = []playlist.Signature{sig}
	signed, _ := json.Marshal(pl)

	other, _, _ := ed25519.GenerateKey(nil)
	hint, _ := sign.Ed25519DIDKey(other)

	err := Run(signed, Playlist, hint)
	if err == nil || !strings.Contains(err.Error(), "no signature entry matches") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRun_legacy_requiresPubkeyWhenOmitted(t *testing.T) {
	t.Parallel()
	pub, priv, _ := ed25519.GenerateKey(nil)
	pl := playlist.Playlist{
		DPVersion: "1.0.0",
		Title:     "L",
		Items:     []playlist.PlaylistItem{{Source: "https://y"}},
	}
	raw, err := json.Marshal(pl)
	if err != nil {
		t.Fatal(err)
	}
	leg, err := sign.SignLegacyEd25519(raw, priv)
	if err != nil {
		t.Fatal(err)
	}
	pl.Signature = leg
	signed, err := json.Marshal(pl)
	if err != nil {
		t.Fatal(err)
	}

	err = Run(signed, Playlist, "")
	if err == nil || !strings.Contains(err.Error(), "legacy") {
		t.Fatalf("expected legacy pubkey message, got %v", err)
	}
	kid, _ := sign.Ed25519DIDKey(pub)
	if err := Run(signed, Playlist, kid); err != nil {
		t.Fatal(err)
	}
}
