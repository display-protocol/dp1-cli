package create

import (
	"strings"
	"time"
)

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
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
