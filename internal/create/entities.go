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

func promptEntityWithTracker(tracker *ask.FieldTracker, scope string) (identity.Entity, error) {
	tracker.Display()
	name, err := ask.LineWithTracker(tracker, scope+" entity name", "", false, nonEmpty())
	if err != nil {
		return identity.Entity{}, err
	}
	tracker.Display()
	key, err := ask.LineWithTracker(tracker, scope+" entity key did:… ", "", false, fields.DID)
	if err != nil {
		return identity.Entity{}, err
	}
	tracker.Display()
	url, err := ask.LineWithTracker(tracker, scope+" entity url (optional)", "", true, fields.URIEmptyOK)
	if err != nil {
		return identity.Entity{}, err
	}
	ent := buildEntity(name, key, url)
	return ent, nil
}

func promptPlaylistURIsWithTracker(tracker *ask.FieldTracker, prompt string) ([]string, error) {
	var rows []string
	uriNum := 1
	for {
		l := `"` + prompt + `" (URI)`
		addEmpty := len(rows) > 0
		if addEmpty {
			l += "; leave empty when done"
		}
		if uriNum > 1 {
			l = fmt.Sprintf("[URI %d] %s", uriNum, l)
		}
		tracker.Display()
		s, err := ask.LineWithTracker(tracker, l, "", addEmpty, func(x string) error {
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
		uriNum++
	}
	return ensurePlaylistURIs(rows)
}
