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

func validateChannelSummary(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if len(s) > 2000 {
		return fmt.Errorf("too long")
	}
	return nil
}

// Channel prompts for a DP-1 channel document (without signatures — run `channel sign`).
func Channel() (*ch.Channel, error) {
	tracker := ask.NewFieldTracker()

	tracker.Display()
	version, err := ask.LineWithTracker(tracker, "Extension version semver", "0.1.0", false, fields.SemVer)
	if err != nil {
		return nil, err
	}

	tracker.Display()
	idHint, err := ask.LineWithTracker(tracker, "Channel id UUID v4 (optional, empty = generate)", "", true, fields.UUIDv4EmptyOK)
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
	if strings.TrimSpace(slug) == "" {
		slug = slugFromTitle(title)
		tracker.UpdateLastField(slug)
	}

	tracker.Display()
	created, err := ask.LineWithTracker(tracker, "Created RFC3339 (optional, empty = now)", "", true, nil)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(created) == "" {
		created = nowRFC3339()
		tracker.UpdateLastField(created)
	}

	plURIs, err := promptPlaylistURIsWithTracker(tracker, "Channel playlist URIs")
	if err != nil {
		return nil, err
	}

	var curators []identity.Entity
	curatorNum := 1
	for {
		tracker.Display()
		add, err := ask.ConfirmWithTracker(tracker, "Add a curator entity?", false)
		if err != nil {
			return nil, err
		}
		if !add {
			break
		}
		ent, err := promptEntityWithTracker(tracker, fmt.Sprintf(`Curator %d`, curatorNum))
		if err != nil {
			return nil, err
		}
		curators = append(curators, ent)
		curatorNum++
	}

	tracker.Display()
	pubYes, err := ask.ConfirmWithTracker(tracker, "Set publisher identity?", false)
	if err != nil {
		return nil, err
	}
	var pub *identity.Entity
	if pubYes {
		ent, err := promptEntityWithTracker(tracker, "Publisher")
		if err != nil {
			return nil, err
		}
		pub = &ent
	}

	tracker.Display()
	sum, err := ask.LineWithTracker(tracker, `Summary text (optional, 1–2000 chars if set)`, "", true, validateChannelSummary)
	if err != nil {
		return nil, err
	}

	tracker.Display()
	cover, err := ask.LineWithTracker(tracker, "coverImage URI (optional)", "", true, fields.URIEmptyOK)
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
