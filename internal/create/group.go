package create

import (
	"strings"

	"github.com/display-protocol/dp1-go/playlistgroup"

	"github.com/display-protocol/dp1-cli/internal/ask"
	"github.com/display-protocol/dp1-cli/internal/fields"
	"github.com/display-protocol/dp1-cli/internal/uuid"
)

// Group prompts for a playlist-group document (without signatures — run `group sign`).
func Group() (*playlistgroup.Group, error) {
	idHint, err := ask.Line("Group id UUID v4 (optional)", "", true, fields.UUIDv4EmptyOK)
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

	title, err := ask.Line("Title", "", false, nonEmpty())
	if err != nil {
		return nil, err
	}

	slug, err := ask.Line("Slug (optional)", "", true, fields.SlugEmptyOK)
	if err != nil {
		return nil, err
	}

	curator, err := ask.Line("Curator name (optional)", "", true, nil)
	if err != nil {
		return nil, err
	}

	sum, err := ask.Line("Summary (optional)", "", true, nil)
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

	plURIs, err := promptPlaylistURIs("Playlist URIs in this group")
	if err != nil {
		return nil, err
	}

	cover, err := ask.Line("coverImage URI (optional)", "", true, fields.URIEmptyOK)
	if err != nil {
		return nil, err
	}

	return &playlistgroup.Group{
		ID:         id,
		Slug:       slug,
		Title:      title,
		Curator:    curator,
		Summary:    sum,
		Playlists:  plURIs,
		Created:    created,
		CoverImage: cover,
	}, nil
}
