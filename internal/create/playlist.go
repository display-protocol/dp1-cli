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
	verDefault := "1.1.0"
	ver, err := ask.Line("dpVersion", verDefault, false, fields.SemVer)
	if err != nil {
		return nil, err
	}

	idHint, err := ask.Line("Playlist id UUID v4 (optional, empty = generate)", "", true, fields.UUIDv4EmptyOK)
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

	created, err := ask.Line("Created RFC3339 (optional, empty = now)", "", true, nil)
	if err != nil {
		return nil, err
	}
	if created == "" {
		created = nowRFC3339()
	}

	var defaults *pl.Defaults
	ok, err := ask.Confirm("Configure playlist defaults (display/license/duration)?", false)
	if err != nil {
		return nil, err
	}
	if ok {
		defaults = &pl.Defaults{}

		ok2, err := ask.Confirm("Set defaults.display?", false)
		if err != nil {
			return nil, err
		}
		if ok2 {
			disp, err := promptDisplayPrefs()
			if err != nil {
				return nil, err
			}
			defaults.Display = disp
		}

		defLic, err := ask.Line(`Defaults license ('open'|'closed') (optional)`, "", true, func(s string) error {
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

		dur, err := ask.FloatEmptyOK("Default duration seconds")
		if err != nil {
			return nil, err
		}
		defaults.Duration = dur
		if defaults.Display == nil && defaults.License == "" && defaults.Duration == nil {
			defaults = nil
		}
	}

	items, err := promptPlaylistItems()
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

func promptPlaylistItems() ([]pl.PlaylistItem, error) {
	var items []pl.PlaylistItem
	for {
		label := `Item artwork "source" URI`
		var allowEmpty bool
		if len(items) > 0 {
			allowEmpty = true
			label += " ; leave empty to finish items"
		}
		source, err := ask.Line(label, "", allowEmpty, func(s string) error {
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

		itemTitle, err := ask.Line("Item title (optional)", "", true, nil)
		if err != nil {
			return nil, err
		}

		itemIDHint, err := ask.Line("Item id UUID v4 (optional)", "", true, fields.UUIDv4EmptyOK)
		if err != nil {
			return nil, err
		}
		itemSlug, err := ask.Line("Item slug (optional)", "", true, fields.SlugEmptyOK)
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
		} else {
			return nil, err
		}

		duration, err := ask.FloatEmptyOK("Duration seconds")
		if err != nil {
			return nil, err
		}
		it.Duration = duration

		license, err := ask.Line(`Item license ('open'|'closed') (optional)`, "", true, func(s string) error {
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

		ref, err := ask.Line("Metadata ref URI (optional)", "", true, fields.URIEmptyOK)
		if err != nil {
			return nil, err
		}
		if ref != "" {
			it.Ref = ref
		}

		ad, err := ask.Confirm("Configure item.display?", false)
		if err != nil {
			return nil, err
		}
		if ad {
			disp, err := promptDisplayPrefs()
			if err != nil {
				return nil, err
			}
			it.Display = disp
		}

		rb, err := ask.Confirm("Configure item.repro?", false)
		if err != nil {
			return nil, err
		}
		if rb {
			r, err := promptReproBlock()
			if err != nil {
				return nil, err
			}
			it.Repro = r
		}

		pb, err := ask.Confirm("Configure item.provenance?", false)
		if err != nil {
			return nil, err
		}
		if pb {
			p, err := promptProvenance()
			if err != nil {
				return nil, err
			}
			it.Provenance = p
		}

		ovRaw, err := ask.Line(`Item "override" raw JSON object (optional)`, "", true, func(s string) error {
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
	}
	return items, nil
}

func promptDisplayPrefs() (*pl.DisplayPrefs, error) {
	_, scaling, err := ask.Select(`display.scaling`,
		[]string{"fit", "fill", "stretch", "auto"})
	if err != nil {
		return nil, err
	}
	bg, err := ask.Line("display.background (#RRGGBB or transparent) [optional, default #000000]", "", true,
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
	autoplayYes, err := ask.Confirm(`display.autoplay default true`, true)
	if err != nil {
		return nil, err
	}
	loopYes, err := ask.Confirm(`display.loop default true`, true)
	if err != nil {
		return nil, err
	}
	iYes, err := ask.Confirm("Configure display.interaction (keyboard/mouse)?", false)
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
		rawKeys, err := ask.Line(`Interaction keyboard codes (comma-separated, optional)`, "", true, nil)
		if err != nil {
			return nil, err
		}
		mp, err := ask.Confirm("Enable mouse.click?", false)
		if err != nil {
			return nil, err
		}
		ms, err := ask.Confirm("Enable mouse.scroll?", false)
		if err != nil {
			return nil, err
		}
		md, err := ask.Confirm("Enable mouse.drag?", false)
		if err != nil {
			return nil, err
		}
		mh, err := ask.Confirm("Enable mouse.hover?", false)
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

func promptReproBlock() (*pl.ReproBlock, error) {
	r := &pl.ReproBlock{}
	chromium, err := ask.Line(`repro.engineVersion.chromium version (optional)`, "", true, nil)
	if err != nil {
		return nil, err
	}
	if chromium != "" {
		r.EngineVersion = map[string]string{"chromium": chromium}
	}
	r.Seed, err = ask.Line("repro.seed (hex 0x… , optional)", "", true, func(s string) error {
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
	hashLine, err := ask.Line(`repro.assetsSHA256 hashes (comma-separated 64-char hex each, optional)`, "", true, nil)
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
	fhSha, err := ask.Line(`repro.frameHash.sha256 (optional 64-char hex)`, "", true, func(s string) error {
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
	ph, err := ask.Line(`repro.frameHash.phash (optional 0x hex)`, "", true, func(s string) error {
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

func promptProvenance() (*pl.ProvenanceBlock, error) {
	_, t, err := ask.Select("provenance.type", []string{
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
		ch, err := ask.Line(`contract.chain (evm|tezos|bitmark|other)`, "", false, func(s string) error {
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
		st, err := ask.Line(`contract.standard (erc721|erc1155|fa2|other)`, "", true, func(s string) error {
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
		c.Address, err = ask.Line("contract.address (optional)", "", true, nil)
		if err != nil {
			return nil, err
		}
		if si, err := ask.IntEmptyOK("contract.seriesId integer (optional)"); err != nil {
			return nil, err
		} else if si != nil {
			c.SeriesID = si
		}
		c.TokenID, err = ask.Line("contract.tokenId (optional)", "", true, nil)
		if err != nil {
			return nil, err
		}
		u, err := ask.Line(`contract.uri (optional)`, "", true, fields.URIEmptyOK)
		if err != nil {
			return nil, err
		}
		c.URI = u
		meta, err := ask.Line("contract.metaHash (optional)", "", true, nil)
		if err != nil {
			return nil, err
		}
		c.MetaHash = meta
		p.Contract = c
	}

	depsDone, err := ask.Confirm("Add provenance.dependencies entries?", false)
	if err != nil {
		return nil, err
	}
	if depsDone {
		for {
			dchain, err := ask.Line(`dep.chain (leave empty when done)`, "", true, func(s string) error {
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
			std, err := ask.Line(`dep.standard`, "", false, func(s string) error {
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
			du, err := ask.Line(`dep.uri`, "", false, fields.URI)
			if err != nil {
				return nil, err
			}
			p.Dependencies = append(p.Dependencies, pl.ProvenanceDep{
				Chain:    dchain,
				Standard: std,
				URI:      du,
			})
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
