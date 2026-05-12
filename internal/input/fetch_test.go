package input

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadSource_file(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.json")
	want := []byte(`{"dpVersion":"1.1.0","title":"x","items":[{"source":"https://a"}]}`)
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSource(path)
	if err != nil {
		t.Fatalf("ReadSource: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReadSource_base64Inline(t *testing.T) {
	t.Parallel()
	plain := []byte(`{"x":true}`)
	b64 := base64.StdEncoding.EncodeToString(plain)
	got, err := ReadSource(b64)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plain) {
		t.Fatalf("got %q want %q", got, plain)
	}
}

func TestReadSource_HTTP(t *testing.T) {
	t.Parallel()
	const payload = `{"dpVersion":"1.1.0","title":"t","items":[{"source":"https://x"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("expected User-Agent header")
		}
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	got, err := ReadSource(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != payload {
		t.Fatalf("got %q want %q", got, payload)
	}
}

func TestReadSource_unsupportedURLScheme(t *testing.T) {
	t.Parallel()
	_, err := ReadSource("ftp://example.com/x")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unsupported URL scheme") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestReadSource_notFound(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "nope.json")
	_, err := ReadSource(path)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadSource_garbage(t *testing.T) {
	t.Parallel()
	_, err := ReadSource("### not a file url or base64 ###")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchURL_nonOK(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no", http.StatusNotFound)
	}))
	defer srv.Close()
	_, err := fetchURL(srv.URL)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("got %v", err)
	}
}

func TestReadSource_trimSpace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.json")
	content := []byte(`{}`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSource("  " + path + "\t")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "{}" {
		t.Fatalf("got %s", got)
	}
}

func TestFetchURL_serverClosed(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	_, err := fetchURL(url)
	if err == nil {
		t.Fatal("expected error after server closed")
	}
}
