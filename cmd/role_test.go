package cmd

import (
	"testing"

	pl "github.com/display-protocol/dp1-go/playlist"
)

func TestParseMultiSigRole_valid(t *testing.T) {
	cases := []struct{ in, want string }{
		{"curator", pl.RoleCurator},
		{"  FEED ", pl.RoleFeed},
		{"Agent", pl.RoleAgent},
		{pl.RoleInstitution, pl.RoleInstitution},
		{"LICENSOR", pl.RoleLicensor},
	}
	for _, c := range cases {
		got, err := parseMultiSigRole(c.in)
		if err != nil {
			t.Fatalf("%q: %v", c.in, err)
		}
		if got != c.want {
			t.Fatalf("%q: got %q want %q", c.in, got, c.want)
		}
	}
}

func TestParseMultiSigRole_invalid(t *testing.T) {
	for _, in := range []string{"", "owner", "publisher"} {
		if _, err := parseMultiSigRole(in); err == nil {
			t.Fatalf("expected error for %q", in)
		}
	}
}
