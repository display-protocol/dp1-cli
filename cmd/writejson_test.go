package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestMarshallPretty_roundTrip(t *testing.T) {
	v := map[string]any{"a": 1, "b": "<tag>"}
	raw, err := marshallPretty(v)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("<tag>")) || bytes.Contains(raw, []byte("\\u003c")) {
		t.Fatalf("unexpected escaping: %s", raw)
	}
}

func TestWriteJSONDocument_writesFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.json")
	doc := struct {
		X int `json:"x"`
	}{X: 42}
	if err := writeJSONDocument(dest, doc, 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(`"x": 42`)) || b[len(b)-1] != '\n' {
		t.Fatalf("file content: %q", b)
	}
}

func TestWriteRawDocument_stdout(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)
	data := []byte(`{"ok":true}` + "\n")
	if err := writeRawDocument(cmd, "-", data); err != nil {
		t.Fatal(err)
	}
	if buf.String() != string(data) {
		t.Fatalf("stdout: %q", buf.String())
	}
}

func TestWriteRawDocument_writesFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "signed.json")
	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	payload := []byte(`{"ok":true}` + "\n")
	if err := writeRawDocument(cmd, dest, payload); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("got %q", got)
	}
}
