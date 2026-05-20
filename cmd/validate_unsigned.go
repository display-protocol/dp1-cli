package cmd

import (
	"encoding/json"
	"fmt"

	dp1 "github.com/display-protocol/dp1-go"
	dp1ch "github.com/display-protocol/dp1-go/extension/channels"
	pl "github.com/display-protocol/dp1-go/playlist"
	"github.com/display-protocol/dp1-go/playlistgroup"
	"github.com/spf13/cobra"

	"github.com/display-protocol/dp1-cli/internal/output"
	"github.com/display-protocol/dp1-cli/internal/validateerr"
)

// validateAllowUnsigned is bound on playlist, group, and channel validate subcommands.
var validateAllowUnsigned bool

const unsignedDraftValidateMsg = "valid except missing signature (run sign next)"

func init() {
	flag := func(c *cobra.Command) {
		c.Flags().BoolVar(&validateAllowUnsigned, "allow-unsigned", false,
			"Accept documents that fail schema only because signatures (or legacy signature) are missing or empty")
	}
	flag(playlistValidateCmd)
	flag(groupValidateCmd)
	flag(channelValidateCmd)
}

// validatePlaylist parses and validates playlist JSON. When allowUnsigned is set and schema
// validation fails only because signatures are missing or empty, the document is still
// accepted as an unsigned draft.
func validatePlaylist(data []byte, allowUnsigned bool) (output.ValidateOK, error) {
	p, err := dp1.ParseAndValidatePlaylist(data)
	if err == nil {
		return output.ValidateOK{
			Resource:  "playlist",
			DPVersion: p.DPVersion,
			ID:        p.ID,
			Title:     p.Title,
		}, nil
	}
	if !allowUnsigned || !validateerr.OnlyMissingSignature(err) {
		return output.ValidateOK{}, err
	}
	var draft pl.Playlist
	if err := json.Unmarshal(data, &draft); err != nil {
		return output.ValidateOK{}, fmt.Errorf("decode playlist (unsigned draft): %w", err)
	}
	return output.ValidateOK{
		Resource:      "playlist",
		DPVersion:     draft.DPVersion,
		ID:            draft.ID,
		Title:         draft.Title,
		UnsignedDraft: true,
		Message:       unsignedDraftValidateMsg,
	}, nil
}

func validateGroup(data []byte, allowUnsigned bool) (output.ValidateOK, error) {
	g, err := dp1.ParseAndValidatePlaylistGroup(data)
	if err == nil {
		return output.ValidateOK{
			Resource: "playlist-group",
			ID:       g.ID,
			Title:    g.Title,
		}, nil
	}
	if !allowUnsigned || !validateerr.OnlyMissingSignature(err) {
		return output.ValidateOK{}, err
	}
	var draft playlistgroup.Group
	if err := json.Unmarshal(data, &draft); err != nil {
		return output.ValidateOK{}, fmt.Errorf("decode playlist-group (unsigned draft): %w", err)
	}
	return output.ValidateOK{
		Resource:      "playlist-group",
		ID:            draft.ID,
		Title:         draft.Title,
		UnsignedDraft: true,
		Message:       unsignedDraftValidateMsg,
	}, nil
}

func validateChannel(data []byte, allowUnsigned bool) (output.ValidateOK, error) {
	ch, err := dp1.ParseAndValidateChannel(data)
	if err == nil {
		return output.ValidateOK{
			Resource: "channel",
			Version:  ch.Version,
			ID:       ch.ID,
			Title:    ch.Title,
		}, nil
	}
	if !allowUnsigned || !validateerr.OnlyMissingSignature(err) {
		return output.ValidateOK{}, err
	}
	var draft dp1ch.Channel
	if err := json.Unmarshal(data, &draft); err != nil {
		return output.ValidateOK{}, fmt.Errorf("decode channel (unsigned draft): %w", err)
	}
	return output.ValidateOK{
		Resource:      "channel",
		Version:       draft.Version,
		ID:            draft.ID,
		Title:         draft.Title,
		UnsignedDraft: true,
		Message:       unsignedDraftValidateMsg,
	}, nil
}
