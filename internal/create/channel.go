package create

import (
	"fmt"
	"strings"

	ch "github.com/display-protocol/dp1-go/extension/channels"
	"github.com/display-protocol/dp1-go/extension/identity"

	"github.com/display-protocol/dp1-cli/internal/ask"
	"github.com/display-protocol/dp1-cli/internal/fields"
	"github.com/display-protocol/dp1-cli/internal/uuid"
)

// Channel prompts for a DP-1 channel document (without signatures — run `channel sign`).
func Channel() (*ch.Channel, error) {
	idHint, err := ask.Line("Channel id UUID v4 (optional)", "", true, fields.UUIDv4EmptyOK)
	if err != nil {
		return nil, err
	}
	id := strings.TrimSpace(idHint)
	if id == "" {
		id, err = uuid.NewV4()
		if err != nil {
			return nil, err
		}
	}

	slug, err := ask.Line("Slug (required, lowercase hyphenated)", "", false, fields.Slug)
	if err != nil {
		return nil, err
	}

	title, err := ask.Line("Title", "", false, nonEmpty())
	if err != nil {
		return nil, err
	}

	version, err := ask.Line("Extension version semver", "0.1.0", false, fields.SemVer)
	if err != nil {
		return nil, err
	}

	created, err := ask.Line("Created RFC3339 (optional)", "", true, nil)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(created) == "" {
		created = nowRFC3339()
	}

	plURIs, err := promptPlaylistURIs("Channel playlist URIs")
	if err != nil {
		return nil, err
	}

	var curators []identity.Entity
	for {
		add, err := ask.Confirm("Add a curator entity?", false)
		if err != nil {
			return nil, err
		}
		if !add {
			break
		}
		ent, err := promptEntity(`Curator`)
		if err != nil {
			return nil, err
		}
		curators = append(curators, ent)
	}

	pubYes, err := ask.Confirm("Set publisher identity?", false)
	if err != nil {
		return nil, err
	}
	var pub *identity.Entity
	if pubYes {
		ent, err := promptEntity("Publisher")
		if err != nil {
			return nil, err
		}
		pub = &ent
	}

	sum, err := ask.Line(`Summary text (optional, 1–2000 chars if set)`, "", true, func(s string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		if len(s) > 2000 {
			return fmt.Errorf("too long")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	cover, err := ask.Line("coverImage URI (optional)", "", true, fields.URIEmptyOK)
	if err != nil {
		return nil, err
	}

	return &ch.Channel{
		ID:         id,
		Slug:       slug,
		Title:      title,
		Version:    version,
		Created:    created,
		Playlists:  plURIs,
		Curators:   curators,
		Publisher:  pub,
		Summary:    sum,
		CoverImage: cover,
	}, nil
}
