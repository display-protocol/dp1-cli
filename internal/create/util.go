package create

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/display-protocol/dp1-cli/internal/fields"
	"github.com/display-protocol/dp1-cli/internal/uuid"
)

const (
	slugRandChars = "abcdefghijklmnopqrstuvwxyz0123456789"

	// uuidSlugSuffixHexLen is appended after a hyphen when a title-derived stem exists
	// (last block of UUID v4, lowercase hex).
	uuidSlugSuffixHexLen = 12

	// maxGeneratedSlugLen caps derived slugs (dp1-go has no slug maxLength; this is UX / URL ergonomics).
	// With hyphen + entropy: title stem budget is maxGeneratedSlugLen - 1 - uuidSlugSuffixHexLen (11 when cap is 24).
	// If you need longer titles in slugs without truncation, bump to 32-48 rather than shortening entropy.
	maxGeneratedSlugLen = 24
)

func slugStemBudget() int {
	return maxGeneratedSlugLen - 1 - uuidSlugSuffixHexLen
}

// truncateSlugStem cuts stem to ≤ max bytes without a trailing hyphen, keeping a valid fields.Slug prefix.
func truncateSlugStem(stem string, max int) string {
	if max <= 0 || stem == "" {
		return ""
	}
	if len(stem) > max {
		stem = stem[:max]
	}
	stem = strings.TrimSuffix(stem, "-")
	for len(stem) > 0 && fields.Slug(stem) != nil {
		stem = stem[:len(stem)-1]
		stem = strings.TrimSuffix(stem, "-")
	}
	return stem
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// slugNormalizeTitle maps title text to a DP-1 slug segment chain (no entropy suffix).
// It returns "" when the title has no usable ASCII letters/digits (caller must not treat "" as a final slug alone).
func slugNormalizeTitle(title string) string {
	title = strings.TrimSpace(title)
	var b strings.Builder
	prevHyphen := true
	for _, r := range strings.ToLower(title) {
		isAZ := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		switch {
		case isAZ || isDigit:
			b.WriteRune(r)
			prevHyphen = false
		case !prevHyphen:
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	s := strings.TrimSuffix(b.String(), "-")
	if s == "" || fields.Slug(s) != nil {
		return ""
	}
	return s
}

// slugSuffixFromUUIDLastBlock returns the last hyphen-separated block of a UUID v4 (12 lowercase hex digits).
func slugSuffixFromUUIDLastBlock() string {
	id, err := uuid.NewV4()
	if err != nil {
		var b [uuidSlugSuffixHexLen / 2]byte
		if _, readErr := rand.Read(b[:]); readErr != nil {
			return "000000000001"
		}
		return hex.EncodeToString(b[:])
	}
	idx := strings.LastIndex(id, "-")
	if idx >= 0 && idx+1 < len(id) {
		last := id[idx+1:]
		if len(last) == uuidSlugSuffixHexLen {
			return last
		}
	}
	var raw [6]byte
	if _, readErr := rand.Read(raw[:]); readErr != nil {
		return "000000000001"
	}
	return hex.EncodeToString(raw[:])
}

// randomAlphanumericSlug returns n characters from [a-z0-9]; used when the title yields no slug stem.
func randomAlphanumericSlug(n int) string {
	if n <= 0 {
		return ""
	}
	rnd := make([]byte, n)
	if _, err := rand.Read(rnd); err != nil {
		out := make([]byte, n)
		for i := range out {
			out[i] = 'a'
		}
		return string(out)
	}
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = slugRandChars[int(rnd[i])%len(slugRandChars)]
	}
	return string(out)
}

// slugFromTitle builds a collision-resistant slug under maxGeneratedSlugLen: truncated title stem plus
// the UUID's last 12-hex block, or an all-random slug when the title cannot be normalized.
func slugFromTitle(title string) string {
	suffix := slugSuffixFromUUIDLastBlock()
	base := slugNormalizeTitle(title)
	if base == "" {
		return randomAlphanumericSlug(maxGeneratedSlugLen)
	}

	maxStem := slugStemBudget()
	if maxStem < 1 {
		return suffix
	}
	if len(base) > maxStem {
		base = truncateSlugStem(base, maxStem)
	}
	if base == "" {
		return suffix
	}
	out := base + "-" + suffix
	if len(out) > maxGeneratedSlugLen {
		return suffix
	}
	return out
}

func splitComma(s string) []string {
	raw := strings.Split(s, ",")
	var out []string
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
