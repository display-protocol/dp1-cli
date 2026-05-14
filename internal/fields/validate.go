package fields

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	// semverPattern matches dotted triples: major.minor.patch (digits only).
	semverPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	// slugPattern matches lowercase hyphenated slugs (no leading/trailing hyphen).
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	// uuidPattern matches RFC 4122 version 4 UUID strings (hex, version nibble 4, RFC variant).
	uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// SemVer validates a non-empty semver string major.minor.patch (e.g. 1.1.0).
func SemVer(s string) error {
	s = strings.TrimSpace(s)
	if !semverPattern.MatchString(s) {
		return fmt.Errorf("must match major.minor.patch (e.g. 1.1.0)")
	}
	return nil
}

// UUIDv4EmptyOK accepts blank input; otherwise requires a UUID version 4 string.
func UUIDv4EmptyOK(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return UUID(s)
}

// UUID requires a UUID version 4 in canonical hyphenated hex form.
func UUID(s string) error {
	s = strings.TrimSpace(s)
	if !uuidPattern.MatchString(s) {
		return fmt.Errorf("must be a UUID v4")
	}
	return nil
}

// SlugEmptyOK accepts blank input; otherwise requires a slug (see Slug).
func SlugEmptyOK(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return Slug(s)
}

// Slug requires lowercase alphanumeric segments separated by single hyphens (no spaces).
func Slug(s string) error {
	s = strings.TrimSpace(s)
	if !slugPattern.MatchString(s) {
		return fmt.Errorf("lowercase alphanumeric segments separated by single hyphens")
	}
	return nil
}

// URI requires an RFC 3986–style URI with a non-empty scheme (https, ipfs, ar, …).
// http/https URLs must be absolute (include a host).
func URI(s string) error {
	s = strings.TrimSpace(s)
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" {
		return fmt.Errorf("must be a URI with a scheme")
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		if u.Host == "" {
			return fmt.Errorf("absolute https URL must include a host")
		}
	}
	return nil
}

// URIEmptyOK accepts blank input; otherwise applies the same rules as URI.
func URIEmptyOK(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return URI(s)
}

// DID performs a shallow check: value must start with "did:" and contain at least
// method and method-specific id segments (did:method:…).
func DID(s string) error {
	s = strings.TrimSpace(s)
	l := strings.ToLower(s)
	if !strings.HasPrefix(l, "did:") || len(strings.Split(l, ":")) < 3 {
		return fmt.Errorf("must start with did:method:…")
	}
	return nil
}
