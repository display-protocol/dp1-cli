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
	tracker := ask.NewFieldTracker()

	tracker.Display()
	idHint, err := ask.LineWithTracker(tracker, "Group id UUID v4 (optional)", "", true, fields.UUIDv4EmptyOK)
	if err != nil {
		return nil, err
	}
	id := strings.TrimSpace(idHint)
	if id == "" {
		id, err = uuid.NewV4()
		if err != nil {
			return nil, err
		}
		tracker.UpdateLastField(id)
	}

	tracker.Display()
	title, err := ask.LineWithTracker(tracker, "Title", "", false, nonEmpty())
	if err != nil {
		return nil, err
	}

	tracker.Display()
	slug, err := ask.LineWithTracker(tracker, "Slug (optional)", "", true, fields.SlugEmptyOK)
	if err != nil {
		return nil, err
	}

	tracker.Display()
	curator, err := ask.LineWithTracker(tracker, "Curator name (optional)", "", true, nil)
	if err != nil {
		return nil, err
	}

	tracker.Display()
	sum, err := ask.LineWithTracker(tracker, "Summary (optional)", "", true, nil)
	if err != nil {
		return nil, err
	}

	tracker.Display()
	created, err := ask.LineWithTracker(tracker, "Created RFC3339 (optional)", "", true, nil)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(created) == "" {
		created = nowRFC3339()
		tracker.UpdateLastField(created)
	}

	plURIs, err := promptPlaylistURIsWithTracker(tracker, "Playlist URIs in this group")
	if err != nil {
		return nil, err
	}

	tracker.Display()
	cover, err := ask.LineWithTracker(tracker, "coverImage URI (optional)", "", true, fields.URIEmptyOK)
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
