package feed

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/display-protocol/dp1-cli/internal/config"
)

func TestResolveCredentials_order(t *testing.T) {
	t.Setenv(EnvURL, "")
	t.Setenv(EnvAPIKey, "")

	_, _, err := ResolveCredentials("", "", config.FeedCfg{})
	if err == nil {
		t.Fatal("expected error when URL missing")
	}

	base, key, err := ResolveCredentials("", "", config.FeedCfg{URL: "https://example.com/v "})
	if err != nil {
		t.Fatal(err)
	}
	if base != "https://example.com/v" {
		t.Fatalf("base URL: %q", base)
	}
	if key != "" {
		t.Fatalf("key: %q", key)
	}

	base, _, err = ResolveCredentials("https://flag.example", "", config.FeedCfg{URL: "https://cfg.example"})
	if err != nil {
		t.Fatal(err)
	}
	if base != "https://flag.example" {
		t.Fatalf("flag URL should win: %q", base)
	}

	t.Setenv(EnvURL, "https://env.example")
	base, _, err = ResolveCredentials("", "", config.FeedCfg{URL: "https://cfg.example"})
	if err != nil {
		t.Fatal(err)
	}
	if base != "https://env.example" {
		t.Fatalf("env URL should beat config: %q", base)
	}

	t.Setenv(EnvAPIKey, "env-secret")
	_, key, err = ResolveCredentials("", "", config.FeedCfg{})
	if err != nil {
		t.Fatal(err)
	}
	if key != "env-secret" {
		t.Fatalf("env key: %q", key)
	}
}

func TestResolveCredentials_flagAPIKeyOverEnv(t *testing.T) {
	t.Setenv(EnvURL, "https://x.example")
	t.Setenv(EnvAPIKey, "from-env")
	base, key, err := ResolveCredentials("", "from-flag", config.FeedCfg{})
	if err != nil {
		t.Fatal(err)
	}
	if base != "https://x.example" || key != "from-flag" {
		t.Fatalf("got %q %q", base, key)
	}
}

func TestClient_Create_success(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/playlists" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method %s", r.Method)
		}
		auth = r.Header.Get("Authorization")
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != `{"ok":true}` {
			t.Fatalf("body %s", b)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"u1"}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient()
	st, body, err := c.Create(context.Background(), srv.URL, Playlist, "k1", []byte(`{"ok":true}`))
	if err != nil {
		t.Fatal(err)
	}
	if st != http.StatusCreated {
		t.Fatalf("status %d", st)
	}
	if auth != "Bearer k1" {
		t.Fatalf("Authorization: %q", auth)
	}
	var doc map[string]string
	if err := json.Unmarshal(body, &doc); err != nil || doc["id"] != "u1" {
		t.Fatalf("body: %s", body)
	}
}

func TestClient_Create_noBearerWhenKeyEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatal("unexpected Authorization header")
		}
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)

	c := NewClient()
	_, _, err := c.Create(context.Background(), srv.URL, Channel, "", []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Create_errorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"validation_error","message":"nope"}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient()
	st, body, err := c.Create(context.Background(), srv.URL, PlaylistGroup, "", []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if st != http.StatusBadRequest {
		t.Fatalf("status %d", st)
	}
	e := ErrorFromResponse(st, body)
	var ae *APIError
	if !errors.As(e, &ae) {
		t.Fatalf("want *APIError, got %T", e)
	}
	if ae.Code != "validation_error" || ae.Message != "nope" {
		t.Fatalf("%#v", ae)
	}
}
