package fields

import "testing"

func TestSemVer(t *testing.T) {
	for _, s := range []string{"1.1.0", " 0.1.2 ", "10.20.30"} {
		if err := SemVer(s); err != nil {
			t.Fatalf("SemVer(%q): %v", s, err)
		}
	}
	for _, s := range []string{"", "1.0", "v1.0.0", "1.0.0.0", "a.b.c"} {
		if err := SemVer(s); err == nil {
			t.Fatalf("SemVer(%q): want error", s)
		}
	}
}

func TestUUID(t *testing.T) {
	valid := "550e8400-e29b-41d4-a716-446655440000"
	if err := UUID(valid); err != nil {
		t.Fatal(err)
	}
	if err := UUIDv4EmptyOK(""); err != nil {
		t.Fatal(err)
	}
	if err := UUIDv4EmptyOK("  "); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"not-a-uuid", "550e8400-e29b-31d4-a716-446655440000", "550e8400-e29b-41d4-c716-446655440000"} {
		if err := UUID(bad); err == nil {
			t.Fatalf("UUID(%q): want error", bad)
		}
	}
}

func TestSlug(t *testing.T) {
	for _, s := range []string{"a", "a-b", "ab-12-cd"} {
		if err := Slug(s); err != nil {
			t.Fatalf("Slug(%q): %v", s, err)
		}
	}
	if err := SlugEmptyOK(""); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"A", "a--b", "-a", "a-", "a_b"} {
		if err := Slug(bad); err == nil {
			t.Fatalf("Slug(%q): want error", bad)
		}
	}
}

func TestURI(t *testing.T) {
	for _, s := range []string{"https://example.com/path", "ipfs://QmX", "ar://foo"} {
		if err := URI(s); err != nil {
			t.Fatalf("URI(%q): %v", s, err)
		}
	}
	if err := URIEmptyOK(""); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"", "ftp", "noturi", "https://"} {
		if err := URI(bad); err == nil {
			t.Fatalf("URI(%q): want error", bad)
		}
	}
}

func TestDID(t *testing.T) {
	for _, s := range []string{"did:key:z6Mk", `did:pkh:eip155:1:0xABC`} {
		if err := DID(s); err != nil {
			t.Fatalf("DID(%q): %v", s, err)
		}
	}
	for _, bad := range []string{"", "did:", "mailto:x@y"} {
		if err := DID(bad); err == nil {
			t.Fatalf("DID(%q): want error", bad)
		}
	}
}
