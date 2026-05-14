package create

import (
	"strings"
	"testing"
)

func TestBuildEntity(t *testing.T) {
	t.Parallel()
	got := buildEntity("Org", " did:key:x ", " https://ex ")
	if got.Name != "Org" {
		t.Fatalf("Name: got %q", got.Name)
	}
	if got.Key != "did:key:x" {
		t.Fatalf("Key: got %q", got.Key)
	}
	if got.URL != "https://ex" {
		t.Fatalf("URL: got %q", got.URL)
	}

	got2 := buildEntity("O", "k", "   ")
	if got2.URL != "" {
		t.Fatalf("empty URL: got %q", got2.URL)
	}
}

func TestEnsurePlaylistURIs(t *testing.T) {
	t.Parallel()
	_, err := ensurePlaylistURIs(nil)
	if err == nil || !strings.Contains(err.Error(), "at least one") {
		t.Fatalf("nil slice: got err %v", err)
	}
	_, err = ensurePlaylistURIs([]string{})
	if err == nil {
		t.Fatal("empty slice: expected error")
	}
	out, err := ensurePlaylistURIs([]string{"https://a"})
	if err != nil || len(out) != 1 || out[0] != "https://a" {
		t.Fatalf("got %v %v", out, err)
	}
}
