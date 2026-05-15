package jsonsign_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"testing"

	dp1 "github.com/display-protocol/dp1-go"
	pl "github.com/display-protocol/dp1-go/playlist"
	"github.com/display-protocol/dp1-go/sign"

	"github.com/display-protocol/dp1-cli/internal/jsonsign"
)

func TestAppendEd25519_playlistMinimal(t *testing.T) {
	unsigned := []byte(`{
  "dpVersion": "1.1.0",
  "title": "fixture",
  "items": [{"source": "https://example.invalid/art.json"}]
}`)

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	got, err := jsonsign.AppendEd25519(unsigned, priv, pl.RoleCurator, "2026-05-01T12:34:56Z", jsonsign.ValidatePlaylist)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dp1.ParseAndValidatePlaylist(got); err != nil {
		t.Fatalf("signed doc invalid: %v", err)
	}
	pub := priv.Public().(ed25519.PublicKey)
	wantKid, err := sign.Ed25519DIDKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Sigs []pl.Signature `json:"signatures"`
	}
	if err := json.Unmarshal(got, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Sigs) != 1 {
		t.Fatalf("signatures length %d", len(env.Sigs))
	}
	if env.Sigs[0].Alg != pl.AlgEd25519 || env.Sigs[0].Kid != wantKid {
		t.Fatalf("unexpected sig %+v want kid %s", env.Sigs[0], wantKid)
	}
}

func TestAppendEd25519_invalidJSON(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	_, err = jsonsign.AppendEd25519([]byte(`not json`), priv, pl.RoleCurator, "2026-05-01T12:34:56Z", jsonsign.ValidatePlaylist)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppendEd25519_rejectsNonEmptyLegacySignature(t *testing.T) {
	doc := []byte(`{
  "dpVersion": "1.1.0",
  "title": "fixture",
  "signature": "ed25519:aa",
  "items": [{"source": "https://example.invalid/a.json"}]
}`)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	_, err = jsonsign.AppendEd25519(doc, priv, pl.RoleCurator, "2026-05-01T12:34:56Z", jsonsign.ValidatePlaylist)
	if err == nil {
		t.Fatal("expected error for legacy signature")
	}
}

func TestAppendEd25519_preservesUnknownTopLevelKeys(t *testing.T) {
	unsigned := []byte(`{
  "dpVersion": "1.1.0",
  "title": "fixture",
  "items": [{"source": "https://example.invalid/art.json"}],
  "xCustom": {"trace": 1}
}`)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	got, err := jsonsign.AppendEd25519(unsigned, priv, pl.RoleCurator, "2026-05-01T12:34:56Z", jsonsign.ValidatePlaylist)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatal(err)
	}
	raw, ok := m["xCustom"]
	if !ok {
		t.Fatal("xCustom key dropped")
	}
	var inner map[string]any
	if err := json.Unmarshal(raw, &inner); err != nil {
		t.Fatal(err)
	}
	if inner["trace"].(float64) != 1 {
		t.Fatalf("xCustom content: %+v", inner)
	}
}

func TestAppendEd25519_appendsSecondSignature(t *testing.T) {
	unsigned := []byte(`{
  "dpVersion": "1.1.0",
  "title": "fixture",
  "items": [{"source": "https://example.invalid/art.json"}]
}`)
	_, k1, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	_, k2, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	once, err := jsonsign.AppendEd25519(unsigned, k1, pl.RoleCurator, "2026-05-01T12:00:00Z", jsonsign.ValidatePlaylist)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := jsonsign.AppendEd25519(once, k2, pl.RoleFeed, "2026-05-01T12:01:00Z", jsonsign.ValidatePlaylist)
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Sigs []pl.Signature `json:"signatures"`
	}
	if err := json.Unmarshal(twice, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Sigs) != 2 {
		t.Fatalf("want 2 signatures, got %d", len(env.Sigs))
	}
	want1, _ := sign.Ed25519DIDKey(k1.Public().(ed25519.PublicKey))
	want2, _ := sign.Ed25519DIDKey(k2.Public().(ed25519.PublicKey))
	if env.Sigs[0].Kid != want1 || env.Sigs[1].Kid != want2 {
		t.Fatalf("kids %+v want %s then %s", env.Sigs, want1, want2)
	}
	if env.Sigs[0].Role != pl.RoleCurator || env.Sigs[1].Role != pl.RoleFeed {
		t.Fatalf("roles %+v", env.Sigs)
	}
}

func TestAppendEd25519_replacesSameKidAndRole(t *testing.T) {
	unsigned := []byte(`{
  "dpVersion": "1.1.0",
  "title": "fixture",
  "items": [{"source": "https://example.invalid/art.json"}]
}`)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	firstTS := "2026-05-01T12:00:00Z"
	secondTS := "2026-05-01T13:00:00Z"
	once, err := jsonsign.AppendEd25519(unsigned, priv, pl.RoleCurator, firstTS, jsonsign.ValidatePlaylist)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := jsonsign.AppendEd25519(once, priv, pl.RoleCurator, secondTS, jsonsign.ValidatePlaylist)
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Sigs []pl.Signature `json:"signatures"`
	}
	if err := json.Unmarshal(twice, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Sigs) != 1 {
		t.Fatalf("want 1 signature after replace, got %d", len(env.Sigs))
	}
	if env.Sigs[0].Ts != secondTS {
		t.Fatalf("want replaced ts %q, got %q", secondTS, env.Sigs[0].Ts)
	}
	if env.Sigs[0].Role != pl.RoleCurator {
		t.Fatalf("role: %+v", env.Sigs[0])
	}
}

func TestAppendEd25519_sameKidDifferentRolesKeepsBoth(t *testing.T) {
	unsigned := []byte(`{
  "dpVersion": "1.1.0",
  "title": "fixture",
  "items": [{"source": "https://example.invalid/art.json"}]
}`)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	once, err := jsonsign.AppendEd25519(unsigned, priv, pl.RoleCurator, "2026-05-01T12:00:00Z", jsonsign.ValidatePlaylist)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := jsonsign.AppendEd25519(once, priv, pl.RoleFeed, "2026-05-01T12:01:00Z", jsonsign.ValidatePlaylist)
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Sigs []pl.Signature `json:"signatures"`
	}
	if err := json.Unmarshal(twice, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Sigs) != 2 {
		t.Fatalf("want 2 signatures, got %d", len(env.Sigs))
	}
}

func TestAppendEd25519_channelMinimal(t *testing.T) {
	unsigned := []byte(`{
  "id": "11111111-1111-4111-8111-111111111111",
  "slug": "fixture-channel",
  "title": "Fixture",
  "version": "0.1.0",
  "created": "2026-05-01T12:00:00Z",
  "playlists": ["https://example.invalid/playlist.json"]
}`)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	got, err := jsonsign.AppendEd25519(unsigned, priv, pl.RoleCurator, "2026-05-01T12:34:56Z", jsonsign.ValidateChannel)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dp1.ParseAndValidateChannel(got); err != nil {
		t.Fatalf("signed channel invalid: %v", err)
	}
}

func TestAppendEd25519_playlistGroupMinimal(t *testing.T) {
	unsigned := []byte(`{
  "id": "22222222-2222-4222-8222-222222222222",
  "title": "Exhibition",
  "playlists": ["https://example.invalid/p.json"],
  "created": "2026-05-01T12:00:00Z"
}`)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	got, err := jsonsign.AppendEd25519(unsigned, priv, pl.RoleCurator, "2026-05-01T12:34:56Z", jsonsign.ValidatePlaylistGroup)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dp1.ParseAndValidatePlaylistGroup(got); err != nil {
		t.Fatalf("signed group invalid: %v", err)
	}
}
