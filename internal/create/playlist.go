package create

import (
	"encoding/json"
	"fmt"
	"strings"

	pl "github.com/display-protocol/dp1-go/playlist"

	"github.com/display-protocol/dp1-cli/internal/ask"
	"github.com/display-protocol/dp1-cli/internal/fields"
	"github.com/display-protocol/dp1-cli/internal/uuid"
)

// Playlist interactively constructs a playlist document (without signatures — run `playlist sign`).
func Playlist() (*pl.Playlist, error) {
	tracker := ask.NewFieldTracker()

	verDefault := "1.1.0"
	tracker.Display()
	ver, err := ask.LineWithTracker(tracker, "dpVersion", verDefault, false, fields.SemVer)
	if err != nil {
		return nil, err
	}

	tracker.Display()
	idHint, err := ask.LineWithTracker(tracker, "Playlist id UUID v4 (optional, empty = generate)", "", true, fields.UUIDv4EmptyOK)
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
	if created == "" {
		created = nowRFC3339()
		tracker.UpdateLastField(created)
	}

	var defaults *pl.Defaults
	tracker.Display()
	ok, err := ask.ConfirmWithTracker(tracker, "Configure playlist defaults (display/license/duration)?", false)
	if err != nil {
		return nil, err
	}
	if ok {
		defaults = &pl.Defaults{}

		tracker.Display()
		ok2, err := ask.ConfirmWithTracker(tracker, "Set defaults.display?", false)
		if err != nil {
			return nil, err
		}
		if ok2 {
			disp, err := promptDisplayPrefs(tracker)
			if err != nil {
				return nil, err
			}
			defaults.Display = disp
		}

		tracker.Display()
		defLic, err := ask.LineWithTracker(tracker, `Defaults license ('open'|'closed') (optional)`, "", true, func(s string) error {
			s = strings.TrimSpace(s)
			if s == "" {
				return nil
			}
			if s != "open" && s != "closed" {
				return fmt.Errorf(`use "open" or "closed"`)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		if defLic != "" {
			defaults.License = defLic
		}

		tracker.Display()
		dur, err := ask.FloatEmptyOKWithTracker(tracker, "Default duration seconds")
		if err != nil {
			return nil, err
		}
		defaults.Duration = dur
		if defaults.Display == nil && defaults.License == "" && defaults.Duration == nil {
			defaults = nil
		}
	}

	items, err := promptPlaylistItems(tracker)
	if err != nil {
		return nil, err
	}

	p := &pl.Playlist{
		DPVersion: ver,
		ID:        id,
		Title:     title,
		Slug:      slug,
		Created:   created,
		Defaults:  defaults,
		Items:     items,
	}
	return p, nil
}

func promptPlaylistItems(tracker *ask.FieldTracker) ([]pl.PlaylistItem, error) {
	var items []pl.PlaylistItem
	itemNum := 1
	for {
		label := `Item artwork "source" URI`
		var allowEmpty bool
		if len(items) > 0 {
			allowEmpty = true
			label += " ; leave empty to finish items"
		}
		tracker.Display()
		source, err := ask.LineWithTracker(tracker, fmt.Sprintf("[Item %d] %s", itemNum, label), "", allowEmpty, func(s string) error {
			if strings.TrimSpace(s) == "" {
				return nil
			}
			return fields.URI(s)
		})
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(source) == "" {
			if len(items) == 0 {
				return nil, fmt.Errorf("playlist needs at least one item")
			}
			break
		}

		tracker.Display()
		itemTitle, err := ask.LineWithTracker(tracker, fmt.Sprintf("[Item %d] Title (optional)", itemNum), "", true, nil)
		if err != nil {
			return nil, err
		}

		tracker.Display()
		itemIDHint, err := ask.LineWithTracker(tracker, fmt.Sprintf("[Item %d] id UUID v4 (optional)", itemNum), "", true, fields.UUIDv4EmptyOK)
		if err != nil {
			return nil, err
		}
		tracker.Display()
		itemSlug, err := ask.LineWithTracker(tracker, fmt.Sprintf("[Item %d] slug (optional)", itemNum), "", true, fields.SlugEmptyOK)
		if err != nil {
			return nil, err
		}

		it := pl.PlaylistItem{
			Source: source,
			Title:  itemTitle,
			Slug:   itemSlug,
		}
		if strings.TrimSpace(itemIDHint) != "" {
			it.ID = strings.TrimSpace(itemIDHint)
		} else if idN, err := uuid.NewV4(); err == nil {
			it.ID = idN
			tracker.UpdateLastField(idN)
		} else {
			return nil, err
		}

		tracker.Display()
		duration, err := ask.FloatEmptyOKWithTracker(tracker, fmt.Sprintf("[Item %d] Duration seconds", itemNum))
		if err != nil {
			return nil, err
		}
		it.Duration = duration

		tracker.Display()
		license, err := ask.LineWithTracker(tracker, fmt.Sprintf(`[Item %d] license ('open'|'closed') (optional)`, itemNum), "", true, func(s string) error {
			s = strings.TrimSpace(s)
			if s == "" {
				return nil
			}
			if s != "open" && s != "closed" {
				return fmt.Errorf(`use "open" or "closed"`)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		if license != "" {
			it.License = license
		}

		tracker.Display()
		ref, err := ask.LineWithTracker(tracker, fmt.Sprintf("[Item %d] Metadata ref URI (optional)", itemNum), "", true, fields.URIEmptyOK)
		if err != nil {
			return nil, err
		}
		if ref != "" {
			it.Ref = ref
		}

		tracker.Display()
		ad, err := ask.ConfirmWithTracker(tracker, fmt.Sprintf("[Item %d] Configure item.display?", itemNum), false)
		if err != nil {
			return nil, err
		}
		if ad {
			disp, err := promptDisplayPrefs(tracker)
			if err != nil {
				return nil, err
			}
			it.Display = disp
		}

		tracker.Display()
		rb, err := ask.ConfirmWithTracker(tracker, fmt.Sprintf("[Item %d] Configure item.repro?", itemNum), false)
		if err != nil {
			return nil, err
		}
		if rb {
			r, err := promptReproBlock(tracker)
			if err != nil {
				return nil, err
			}
			it.Repro = r
		}

		tracker.Display()
		pb, err := ask.ConfirmWithTracker(tracker, fmt.Sprintf("[Item %d] Configure item.provenance?", itemNum), false)
		if err != nil {
			return nil, err
		}
		if pb {
			p, err := promptProvenance(tracker)
			if err != nil {
				return nil, err
			}
			it.Provenance = p
		}

		tracker.Display()
		ovRaw, err := ask.LineWithTracker(tracker, fmt.Sprintf(`[Item %d] "override" raw JSON object (optional)`, itemNum), "", true, func(s string) error {
			if strings.TrimSpace(s) == "" {
				return nil
			}
			var m json.RawMessage
			return json.Unmarshal([]byte(strings.TrimSpace(s)), &m)
		})
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(ovRaw) != "" {
			it.Override = json.RawMessage(strings.TrimSpace(ovRaw))
		}

		items = append(items, it)
		itemNum++
	}
	return items, nil
}

func promptDisplayPrefs(tracker *ask.FieldTracker) (*pl.DisplayPrefs, error) {
	tracker.Display()
	_, scaling, err := ask.SelectWithTracker(tracker, `display.scaling`,
		[]string{"fit", "fill", "stretch", "auto"})
	if err != nil {
		return nil, err
	}
	tracker.Display()
	bg, err := ask.LineWithTracker(tracker, "display.background (#RRGGBB or transparent) [optional, default #000000]", "", true,
		func(s string) error {
			s = strings.TrimSpace(s)
			if s == "" {
				return nil
			}
			if s != "transparent" && (len(s) != 7 || s[0] != '#') {
				return fmt.Errorf("#RRGGBB or transparent")
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	tracker.Display()
	autoplayYes, err := ask.ConfirmWithTracker(tracker, `display.autoplay default true`, true)
	if err != nil {
		return nil, err
	}
	tracker.Display()
	loopYes, err := ask.ConfirmWithTracker(tracker, `display.loop default true`, true)
	if err != nil {
		return nil, err
	}
	tracker.Display()
	iYes, err := ask.ConfirmWithTracker(tracker, "Configure display.interaction (keyboard/mouse)?", false)
	if err != nil {
		return nil, err
	}
	d := &pl.DisplayPrefs{
		Scaling: scaling,
	}
	if bg != "" {
		d.Background = bg
	} else {
		d.Background = "#000000"
	}
	t := autoplayYes
	d.Autoplay = &t
	l := loopYes
	d.Loop = &l

	if iYes {
		tracker.Display()
		rawKeys, err := ask.LineWithTracker(tracker, `Interaction keyboard codes (comma-separated, optional)`, "", true, nil)
		if err != nil {
			return nil, err
		}
		tracker.Display()
		mp, err := ask.ConfirmWithTracker(tracker, "Enable mouse.click?", false)
		if err != nil {
			return nil, err
		}
		tracker.Display()
		ms, err := ask.ConfirmWithTracker(tracker, "Enable mouse.scroll?", false)
		if err != nil {
			return nil, err
		}
		tracker.Display()
		md, err := ask.ConfirmWithTracker(tracker, "Enable mouse.drag?", false)
		if err != nil {
			return nil, err
		}
		tracker.Display()
		mh, err := ask.ConfirmWithTracker(tracker, "Enable mouse.hover?", false)
		if err != nil {
			return nil, err
		}
		ip := &pl.InteractionPrefs{
			Mouse: &pl.MousePrefs{Click: mp, Scroll: ms, Drag: md, Hover: mh},
		}
		ip.Keyboard = splitComma(rawKeys)
		d.Interaction = ip
	}
	return d, nil
}

func promptReproBlock(tracker *ask.FieldTracker) (*pl.ReproBlock, error) {
	r := &pl.ReproBlock{}
	tracker.Display()
	chromium, err := ask.LineWithTracker(tracker, `repro.engineVersion.chromium version (optional)`, "", true, nil)
	if err != nil {
		return nil, err
	}
	if chromium != "" {
		r.EngineVersion = map[string]string{"chromium": chromium}
	}
	tracker.Display()
	r.Seed, err = ask.LineWithTracker(tracker, "repro.seed (hex 0x… , optional)", "", true, func(s string) error {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		if len(strings.TrimSpace(s)) < 4 || strings.TrimSpace(s)[:2] != "0x" {
			return fmt.Errorf(`seed must match ^0x[a-f0-9]+$ in schema`)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	tracker.Display()
	hashLine, err := ask.LineWithTracker(tracker, `repro.assetsSHA256 hashes (comma-separated 64-char hex each, optional)`, "", true, nil)
	if err != nil {
		return nil, err
	}
	if hl := strings.TrimSpace(hashLine); hl != "" {
		for _, p := range splitComma(hl) {
			if len(p) != 64 {
				return nil, fmt.Errorf("sha256 hashes must be 64 hex chars")
			}
			r.AssetsSHA256 = append(r.AssetsSHA256, p)
		}
	}
	tracker.Display()
	fhSha, err := ask.LineWithTracker(tracker, `repro.frameHash.sha256 (optional 64-char hex)`, "", true, func(s string) error {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		if len(strings.TrimSpace(s)) != 64 {
			return fmt.Errorf("want 64 hex chars")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	tracker.Display()
	ph, err := ask.LineWithTracker(tracker, `repro.frameHash.phash (optional 0x hex)`, "", true, func(s string) error {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		vv := strings.TrimSpace(s)
		if len(vv) < 3 || vv[:2] != "0x" {
			return fmt.Errorf(`must start with 0x`)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if fhSha != "" || ph != "" {
		r.FrameHash = &pl.FrameHash{}
		if fhSha != "" {
			r.FrameHash.SHA256 = fhSha
		}
		if ph != "" {
			r.FrameHash.Phash = ph
		}
	}
	if r.EngineVersion == nil && strings.TrimSpace(r.Seed) == "" && len(r.AssetsSHA256) == 0 && r.FrameHash == nil {
		return nil, fmt.Errorf("repro block was empty")
	}
	return r, nil
}

func promptProvenance(tracker *ask.FieldTracker) (*pl.ProvenanceBlock, error) {
	tracker.Display()
	_, t, err := ask.SelectWithTracker(tracker, "provenance.type", []string{
		string(pl.ProvenanceOnChain),
		string(pl.ProvenanceSeriesRegistry),
		string(pl.ProvenanceOffChainURI),
	})
	if err != nil {
		return nil, err
	}
	p := &pl.ProvenanceBlock{Type: pl.ProvenanceType(t)}

	needContract := p.Type == pl.ProvenanceOnChain || p.Type == pl.ProvenanceSeriesRegistry

	if needContract {
		c := &pl.ProvenanceContract{}
		tracker.Display()
		ch, err := ask.LineWithTracker(tracker, `contract.chain (evm|tezos|bitmark|other)`, "", false, func(s string) error {
			switch strings.TrimSpace(s) {
			case "evm", "tezos", "bitmark", "other":
				return nil
			default:
				return fmt.Errorf(`invalid`)
			}
		})
		if err != nil {
			return nil, err
		}
		c.Chain = ch
		tracker.Display()
		st, err := ask.LineWithTracker(tracker, `contract.standard (erc721|erc1155|fa2|other)`, "", true, func(s string) error {
			switch strings.TrimSpace(s) {
			case "", "erc721", "erc1155", "fa2", "other":
				return nil
			default:
				return fmt.Errorf(`invalid`)
			}
		})
		if err != nil {
			return nil, err
		}
		c.Standard = st
		tracker.Display()
		c.Address, err = ask.LineWithTracker(tracker, "contract.address (optional)", "", true, nil)
		if err != nil {
			return nil, err
		}
		tracker.Display()
		if si, err := ask.IntEmptyOKWithTracker(tracker, "contract.seriesId integer (optional)"); err != nil {
			return nil, err
		} else if si != nil {
			c.SeriesID = si
		}
		tracker.Display()
		c.TokenID, err = ask.LineWithTracker(tracker, "contract.tokenId (optional)", "", true, nil)
		if err != nil {
			return nil, err
		}
		tracker.Display()
		u, err := ask.LineWithTracker(tracker, `contract.uri (optional)`, "", true, fields.URIEmptyOK)
		if err != nil {
			return nil, err
		}
		c.URI = u
		tracker.Display()
		meta, err := ask.LineWithTracker(tracker, "contract.metaHash (optional)", "", true, nil)
		if err != nil {
			return nil, err
		}
		c.MetaHash = meta
		p.Contract = c
	}

	tracker.Display()
	depsDone, err := ask.ConfirmWithTracker(tracker, "Add provenance.dependencies entries?", false)
	if err != nil {
		return nil, err
	}
	if depsDone {
		depNum := 1
		for {
			tracker.Display()
			dchain, err := ask.LineWithTracker(tracker, fmt.Sprintf(`[Dependency %d] chain (leave empty when done)`, depNum), "", true, func(s string) error {
				if strings.TrimSpace(s) == "" {
					return nil
				}
				switch strings.TrimSpace(s) {
				case "evm", "tezos", "bitmark", "other":
					return nil
				default:
					return fmt.Errorf(`invalid`)
				}
			})
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(dchain) == "" {
				break
			}
			tracker.Display()
			std, err := ask.LineWithTracker(tracker, fmt.Sprintf(`[Dependency %d] standard`, depNum), "", false, func(s string) error {
				switch strings.TrimSpace(s) {
				case "erc721", "erc1155", "fa2", "other":
					return nil
				default:
					return fmt.Errorf(`invalid`)
				}
			})
			if err != nil {
				return nil, err
			}
			tracker.Display()
			du, err := ask.LineWithTracker(tracker, fmt.Sprintf(`[Dependency %d] uri`, depNum), "", false, fields.URI)
			if err != nil {
				return nil, err
			}
			p.Dependencies = append(p.Dependencies, pl.ProvenanceDep{
				Chain:    dchain,
				Standard: std,
				URI:      du,
			})
			depNum++
		}
	}

	return p, nil
}

func nonEmpty() func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("cannot be empty")
		}
		return nil
	}
}
