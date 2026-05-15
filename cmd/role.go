package cmd

import (
	"fmt"
	"strings"

	dp1ch "github.com/display-protocol/dp1-go/extension/channels"
	pl "github.com/display-protocol/dp1-go/playlist"
)

func parseMultiSigRole(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case pl.RoleCurator, pl.RoleFeed, pl.RoleAgent, pl.RoleInstitution, pl.RoleLicensor:
		return s, nil
	case dp1ch.RolePublisher:
		return dp1ch.RolePublisher, nil
	case "":
		return "", fmt.Errorf("empty role")
	default:
		return "", fmt.Errorf("unknown role %q (use curator, feed, agent, institution, licensor, publisher)", s)
	}
}
