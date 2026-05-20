package cmd_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/display-protocol/dp1-cli/cmd"
	"github.com/display-protocol/dp1-cli/internal/config"
)

func TestExecute_keyGenerate_saveConfig_doesNotPrintPrivateKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	defer resetCLIState(t)

	root := cmd.Root
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"key", "generate", "--save-config"})
	t.Cleanup(func() {
		root.SetArgs(nil)
		resetCLIState(t)
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	stdout := buf.String()

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Signing.PrivateKey) != 128 {
		t.Fatalf("expected 128-char expanded private key in config, got len %d", len(cfg.Signing.PrivateKey))
	}
	if strings.Contains(stdout, cfg.Signing.PrivateKey) {
		t.Fatal("stdout must not contain the saved private key")
	}
	if !strings.Contains(stdout, "did:key:") {
		t.Fatalf("stdout should include public did:key, got %q", stdout)
	}
	if !strings.Contains(stdout, "saved signing.private_key") {
		t.Fatalf("stdout should confirm save, got %q", stdout)
	}
}

func TestExecute_keyGenerate_withoutSaveConfig_printsPrivateKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	defer resetCLIState(t)

	root := cmd.Root
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"key", "generate"})
	t.Cleanup(func() {
		root.SetArgs(nil)
		resetCLIState(t)
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	stdout := buf.String()

	if !strings.Contains(stdout, "private key hex") {
		t.Fatalf("stdout should include private key for copy, got %q", stdout)
	}

	cfgPath := filepath.Join(home, ".dp1", "config.yaml")
	if _, err := os.Stat(cfgPath); err == nil {
		cfg, err := config.Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Signing.PrivateKey != "" {
			t.Fatal("generate without --save-config should not write private key to config")
		}
	}
}

func TestExecute_keyGenerate_saveConfig_JSON_omitsPrivateKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	defer resetCLIState(t)

	root := cmd.Root
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"--json", "key", "generate", "--save-config"})
	t.Cleanup(func() {
		root.SetArgs(nil)
		resetCLIState(t)
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	stdout := buf.String()

	var env struct {
		OK                    bool   `json:"ok"`
		PublicDID             string `json:"public_did_key"`
		PrivateKeyHexExpanded string `json:"private_key_hex_expanded"`
		Saved                 bool   `json:"saved"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("stdout %q: %v", stdout, err)
	}
	if !env.OK || !env.Saved || env.PublicDID == "" {
		t.Fatalf("got %#v, want ok with saved and public_did_key", env)
	}
	if env.PrivateKeyHexExpanded != "" {
		t.Fatalf("JSON must omit private_key_hex_expanded when --save-config, got %q", env.PrivateKeyHexExpanded)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Signing.PrivateKey == "" {
		t.Fatal("expected private key saved to config")
	}
}

func TestExecute_keyGenerate_JSON_includesPrivateKeyWithoutSave(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	defer resetCLIState(t)

	root := cmd.Root
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"--json", "key", "generate"})
	t.Cleanup(func() {
		root.SetArgs(nil)
		resetCLIState(t)
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	stdout := buf.String()

	var env struct {
		PrivateKeyHexExpanded string `json:"private_key_hex_expanded"`
		Saved                 bool   `json:"saved"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatal(err)
	}
	if len(env.PrivateKeyHexExpanded) != 128 {
		t.Fatalf("want 128-char private key in JSON, got len %d", len(env.PrivateKeyHexExpanded))
	}
	if env.Saved {
		t.Fatal("saved should be absent/false without --save-config")
	}
}
