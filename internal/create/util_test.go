package create

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/display-protocol/dp1-cli/internal/fields"
)

var (
	reSlugSuffixHex      = regexp.MustCompile(`^[0-9a-f]{12}$`)
	reRandomSlugAlphanum = regexp.MustCompile(`^[a-z0-9]{24}$`)
)

func TestSplitComma(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"a", []string{"a"}},
		{"a,b", []string{"a", "b"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"x,,y", []string{"x", "y"}},
	}
	for _, tt := range tests {
		got := splitComma(tt.in)
		if len(got) != len(tt.want) {
			t.Fatalf("splitComma(%q) len got %d want %d", tt.in, len(got), len(tt.want))
		}
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Fatalf("splitComma(%q)[%d] got %q want %q", tt.in, i, got[i], tt.want[i])
			}
		}
	}
}

func TestTruncateSlugStem(t *testing.T) {
	t.Parallel()
	tests := []struct {
		stem string
		max  int
		want string
	}{
		{"hello-world", 11, "hello-world"},
		{"hello-world", 5, "hello"},
		{"hello-world", 6, "hello"},
		{"hello-world", 7, "hello-w"},
		{"hello-world", 8, "hello-wo"},
		{"multiple-spaces-here", 11, "multiple-sp"},
		{"", 10, ""},
		{"hi", 0, ""},
	}
	for _, tt := range tests {
		got := truncateSlugStem(tt.stem, tt.max)
		if got != tt.want {
			t.Fatalf("truncateSlugStem(%q, %d) got %q want %q", tt.stem, tt.max, got, tt.want)
		}
		if got != "" && fields.Slug(got) != nil {
			t.Fatalf("truncateSlugStem(%q, %d) invalid slug %q", tt.stem, tt.max, got)
		}
	}
}

func TestSlugNormalizeTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		title string
		want  string
	}{
		{"My Channel", "my-channel"},
		{"  Hello World  ", "hello-world"},
		{"slug-ready", "slug-ready"},
		{"Multiple   Spaces Here", "multiple-spaces-here"},
		{"htp-ds", "htp-ds"},
		{"Version 2.0", "version-2-0"},
		{"!!!", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := slugNormalizeTitle(tt.title)
		if got != tt.want {
			t.Fatalf("slugNormalizeTitle(%q) got %q want %q", tt.title, got, tt.want)
		}
		if tt.want != "" && fields.Slug(got) != nil {
			t.Fatalf("slugNormalizeTitle(%q) invalid base %q", tt.title, got)
		}
	}
}

func TestSlugFromTitleShape(t *testing.T) {
	t.Parallel()
	titles := []string{"My Channel", "!!!", "", "hello", "slug-ready"}
	for _, title := range titles {
		for range 32 {
			got := slugFromTitle(title)
			if len(got) > maxGeneratedSlugLen {
				t.Fatalf("slugFromTitle(%q) len %d > cap %d got %q", title, len(got), maxGeneratedSlugLen, got)
			}
			if err := fields.Slug(got); err != nil {
				t.Fatalf("slugFromTitle(%q) invalid slug %q: %v", title, got, err)
			}
			base := slugNormalizeTitle(title)
			if base == "" {
				if strings.Contains(got, "-") {
					t.Fatalf("slugFromTitle(%q) empty base want single segment, got %q", title, got)
				}
				if len(got) != maxGeneratedSlugLen || !reRandomSlugAlphanum.MatchString(got) {
					t.Fatalf("slugFromTitle(%q) want %d-char [a-z0-9]+, got %q", title, maxGeneratedSlugLen, got)
				}
				continue
			}

			if len(got) < uuidSlugSuffixHexLen+2 {
				t.Fatalf("slugFromTitle(%q) too short %q", title, got)
			}
			suffix := got[len(got)-uuidSlugSuffixHexLen:]
			if got[len(got)-uuidSlugSuffixHexLen-1] != '-' {
				t.Fatalf("slugFromTitle(%q) want hyphen before entropy, got %q", title, got)
			}
			stem := got[:len(got)-uuidSlugSuffixHexLen-1]
			if !reSlugSuffixHex.MatchString(suffix) {
				t.Fatalf("slugFromTitle(%q) want 12 hex suffix (UUID tail), got suffix %q", title, suffix)
			}
			if fields.Slug(stem) != nil {
				t.Fatalf("slugFromTitle(%q) invalid stem %q", title, stem)
			}
			if len(stem) > slugStemBudget() {
				t.Fatalf("slugFromTitle(%q) stem len %d > budget %d", title, len(stem), slugStemBudget())
			}
			if stem != base && !strings.HasPrefix(base, stem) {
				t.Fatalf("slugFromTitle(%q) stem %q not prefix of normalized base %q", title, stem, base)
			}
		}
	}
}

func TestSlugFromTitleLongTitleRespectsCap(t *testing.T) {
	t.Parallel()
	title := strings.Repeat("word ", 20) // long normalized stem
	for range 16 {
		got := slugFromTitle(title)
		if len(got) > maxGeneratedSlugLen {
			t.Fatalf("len %d > %d: %q", len(got), maxGeneratedSlugLen, got)
		}
		if err := fields.Slug(got); err != nil {
			t.Fatalf("%q: %v", got, err)
		}
	}
}

func TestSlugFromTitleUniquenessSample(t *testing.T) {
	t.Parallel()
	seen := make(map[string]struct{})
	for range 50 {
		s := slugFromTitle("Same Title Here")
		if _, ok := seen[s]; ok {
			t.Fatalf("duplicate slug %q in sample — entropy collision", s)
		}
		seen[s] = struct{}{}
	}
}

func TestSlugFromTitleUniquenessRandomStem(t *testing.T) {
	t.Parallel()
	seen := make(map[string]struct{})
	for range 80 {
		s := slugFromTitle("!!!")
		if _, ok := seen[s]; ok {
			t.Fatalf("duplicate random stem %q — collision", s)
		}
		seen[s] = struct{}{}
	}
}

func TestNowRFC3339(t *testing.T) {
	t.Parallel()
	s := nowRFC3339()
	if !strings.HasSuffix(s, "Z") && !strings.ContainsAny(s, "+-") {
		t.Fatalf("expected UTC-offset RFC3339, got %q", s)
	}
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("Parse RFC3339: %v", err)
	}
	if d := time.Since(ts); d < -time.Minute || d > time.Minute {
		t.Fatalf("timestamp out of range: %v", ts)
	}
}
