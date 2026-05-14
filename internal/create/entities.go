package create

import (
	"fmt"
	"strings"

	"github.com/display-protocol/dp1-go/extension/identity"

	"github.com/display-protocol/dp1-cli/internal/ask"
	"github.com/display-protocol/dp1-cli/internal/fields"
)

func buildEntity(name, key, url string) identity.Entity {
	ent := identity.Entity{Name: name, Key: strings.TrimSpace(key)}
	if strings.TrimSpace(url) != "" {
		ent.URL = strings.TrimSpace(url)
	}
	return ent
}

func ensurePlaylistURIs(rows []string) ([]string, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("need at least one playlist URI")
	}
	return rows, nil
}

func promptEntity(scope string) (identity.Entity, error) {
	name, err := ask.Line(scope+" entity name", "", false, nonEmpty())
	if err != nil {
		return identity.Entity{}, err
	}
	key, err := ask.Line(scope+" entity key did:… ", "", false, fields.DID)
	if err != nil {
		return identity.Entity{}, err
	}
	url, err := ask.Line(scope+" entity url (optional)", "", true, fields.URIEmptyOK)
	if err != nil {
		return identity.Entity{}, err
	}
	ent := buildEntity(name, key, url)
	return ent, nil
}

func promptPlaylistURIs(prompt string) ([]string, error) {
	var rows []string
	for {
		l := `"` + prompt + `" (URI)`
		addEmpty := len(rows) > 0
		if addEmpty {
			l += "; leave empty when done"
		}
		s, err := ask.Line(l, "", addEmpty, func(x string) error {
			if strings.TrimSpace(x) == "" && addEmpty {
				return nil
			}
			return fields.URI(x)
		})
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(s) == "" {
			break
		}
		rows = append(rows, strings.TrimSpace(s))
	}
	return ensurePlaylistURIs(rows)
}
