package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	_ = r.Close()
	return buf.String()
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = old
	}()
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	_ = r.Close()
	return buf.String()
}

func TestPrintValidateSuccess_JSON(t *testing.T) {
	out := captureStdout(t, func() {
		PrintValidateSuccess(true, ValidateOK{
			Resource:  "playlist",
			DPVersion: "1.1.0",
			ID:        "id-1",
			Title:     "T",
		})
	})
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if got["ok"] != true {
		t.Fatalf("ok: %v", got["ok"])
	}
	if got["resource"] != "playlist" {
		t.Fatalf("resource: %v", got["resource"])
	}
	if got["dpVersion"] != "1.1.0" {
		t.Fatalf("dpVersion: %v", got["dpVersion"])
	}
}

func TestPrintError_JSON(t *testing.T) {
	out := captureStdout(t, func() {
		PrintError(true, ErrorReport{Command: "playlist validate", Error: "boom"})
	})
	var got ErrorReport
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.OK {
		t.Fatal("expected ok false")
	}
	if got.Error != "boom" {
		t.Fatalf("error field: %q", got.Error)
	}
}

func TestPrintVerifySuccess_JSON(t *testing.T) {
	out := captureStdout(t, func() {
		PrintVerifySuccess(true, VerifyOK{Resource: "channel", Mode: "all"})
	})
	var got VerifyOK
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if !got.OK || got.Resource != "channel" {
		t.Fatalf("%+v", got)
	}
}

func TestPrintError_human(t *testing.T) {
	errOut := captureStderr(t, func() {
		PrintError(false, ErrorReport{Command: "x", Error: "y"})
	})
	if !bytes.Contains([]byte(errOut), []byte("y")) {
		t.Fatalf("expected stderr to contain error text: %q", errOut)
	}
}
