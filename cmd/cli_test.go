package cmd_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	dp1 "github.com/display-protocol/dp1-go"
	"github.com/display-protocol/dp1-go/playlist"
	"github.com/display-protocol/dp1-go/sign"
)

func TestCLIPlaylistValidateVerify(t *testing.T) {
	t.Parallel()
	pub, priv, _ := ed25519.GenerateKey(nil)
	pl := playlist.Playlist{
		DPVersion: "1.1.0",
		Title:     "cli-test",
		Items:     []playlist.PlaylistItem{{Source: "https://example.com/a"}},
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
	if _, err := dp1.ParseAndValidatePlaylist(signed); err != nil {
		t.Fatal(err)
	}

	cli := buildCLI(t)
	cmd := exec.Command(cli, "playlist", "validate", "-")
	cmd.Stdin = bytes.NewReader(signed)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("validate: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "valid") {
		t.Fatalf("unexpected output: %s", out)
	}

	cmd = exec.Command(cli, "playlist", "verify", "-")
	cmd.Stdin = bytes.NewReader(signed)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("verify: %v\n%s", err, out)
	}

	kid, err := sign.Ed25519DIDKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command(cli, "playlist", "verify", "--pubkey", kid, "-")
	cmd.Stdin = bytes.NewReader(signed)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("verify pubkey: %v\n%s", err, out)
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
	out := filepath.Join(t.TempDir(), "dp1")
	cmd := exec.Command("go", "build", "-o", out, root)
	if x, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, x)
	}
	return out
}
