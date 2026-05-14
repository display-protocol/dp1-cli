package jsonsign

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"strings"

	dp1 "github.com/display-protocol/dp1-go"
	"github.com/display-protocol/dp1-go/playlist"
	"github.com/display-protocol/dp1-go/sign"
)

// AppendEd25519 signs orig (same bytes hashing expects), preserves unknown JSON keys,
// appends to "signatures", strips legacy "signature", and runs validate on the result.
func AppendEd25519(orig []byte, priv ed25519.PrivateKey, role, tsRFC3339 string, validate func([]byte) error) ([]byte, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(orig, &m); err != nil {
		return nil, fmt.Errorf("decode JSON: %w", err)
	}

	if lv, ok := m["signature"]; ok {
		var leg string
		_ = json.Unmarshal(lv, &leg)
		if strings.TrimSpace(leg) != "" && string(lv) != "null" {
			return nil, fmt.Errorf(`document has legacy "signature" — remove it before adding v1.1 multi-signatures`)
		}
	}

	newSig, err := sign.SignMultiEd25519(orig, priv, role, tsRFC3339)
	if err != nil {
		return nil, err
	}

	var existing []playlist.Signature
	if sg, ok := m["signatures"]; ok && string(sg) != "null" {
		if err := json.Unmarshal(sg, &existing); err != nil {
			return nil, fmt.Errorf("decode signatures: %w", err)
		}
	}
	existing = append(existing, newSig)
	signed, err := json.Marshal(existing)
	if err != nil {
		return nil, err
	}
	m["signatures"] = signed
	delete(m, "signature")

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}

	out := bytes.TrimSuffix(buf.Bytes(), []byte{'\n'})
	newLine := append(out, '\n')

	if validate != nil {
		if err := validate(newLine); err != nil {
			return nil, fmt.Errorf("signed document failed validation: %w", err)
		}
	}
	return newLine, nil
}

// ValidatePlaylist wires [dp1.ParseAndValidatePlaylist].
func ValidatePlaylist(b []byte) error {
	_, err := dp1.ParseAndValidatePlaylist(b)
	return err
}

// ValidateChannel wires [dp1.ParseAndValidateChannel].
func ValidateChannel(b []byte) error {
	_, err := dp1.ParseAndValidateChannel(b)
	return err
}

// ValidatePlaylistGroup wires [dp1.ParseAndValidatePlaylistGroup].
func ValidatePlaylistGroup(b []byte) error {
	_, err := dp1.ParseAndValidatePlaylistGroup(b)
	return err
}
