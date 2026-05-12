package verify

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/display-protocol/dp1-go/playlist"
	"github.com/display-protocol/dp1-go/sign"
	"github.com/ethereum/go-ethereum/common"
)

// Kind selects which Verify*Signatures helper to use for full-document checks.
type Kind int

const (
	Playlist Kind = iota
	PlaylistGroup
	Channel
)

type envelope struct {
	Signatures []playlist.Signature `json:"signatures"`
	Signature  string               `json:"signature"`
}

// Run verifies DP-1 signatures on raw JSON bytes (same bytes used for canonical hashing).
func Run(raw []byte, k Kind, pubkey string) error {
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode document envelope: %w", err)
	}

	hasMulti := len(env.Signatures) > 0
	hasLegacy := strings.TrimSpace(env.Signature) != ""

	if hasMulti {
		if pubkey == "" {
			return verifyAllMulti(raw, k)
		}
		hint, err := ParsePubkeyHint(pubkey)
		if err != nil {
			return err
		}
		return verifyMultiForHint(raw, env.Signatures, hint)
	}

	if hasLegacy {
		if pubkey == "" {
			return fmt.Errorf("document has legacy single signature only; provide --pubkey (Ed25519, 32-byte hex) to verify")
		}
		pub, err := ed25519PubFromFlag(pubkey)
		if err != nil {
			return fmt.Errorf("legacy signature requires Ed25519 public key: %w", err)
		}
		return sign.VerifyLegacyEd25519(raw, env.Signature, pub)
	}

	return sign.ErrNoSignatures
}

func verifyAllMulti(raw []byte, k Kind) error {
	var ok bool
	var err error
	switch k {
	case Playlist:
		ok, _, err = sign.VerifyPlaylistSignatures(raw)
	case PlaylistGroup:
		ok, _, err = sign.VerifyPlaylistGroupSignatures(raw)
	case Channel:
		ok, _, err = sign.VerifyChannelSignatures(raw)
	default:
		return fmt.Errorf("internal: unknown document kind %d", k)
	}
	if err != nil {
		return err
	}
	if !ok {
		return sign.ErrSigInvalid
	}
	return nil
}

func verifyMultiForHint(raw []byte, sigs []playlist.Signature, hint *PubkeyHint) error {
	var matched []playlist.Signature
	for _, s := range sigs {
		ok, err := sigMatchesHint(s, hint)
		if err != nil {
			return err
		}
		if ok {
			matched = append(matched, s)
		}
	}
	if len(matched) == 0 {
		return fmt.Errorf("no signature entry matches the given --pubkey")
	}
	for _, s := range matched {
		if err := sign.VerifyMultiSignature(raw, s); err != nil {
			return fmt.Errorf("signature for kid %q (alg %q): %w", s.Kid, s.Alg, err)
		}
	}
	return nil
}

func sigMatchesHint(s playlist.Signature, hint *PubkeyHint) (bool, error) {
	switch hint.style {
	case hintDIDKey:
		return strings.EqualFold(s.Kid, hint.didKey), nil
	case hintDIDPKH:
		return strings.EqualFold(s.Kid, hint.didPKH), nil
	case hintEd25519Raw:
		if !strings.EqualFold(s.Alg, playlist.AlgEd25519) {
			return false, nil
		}
		expect, err := sign.Ed25519DIDKey(hint.edPub)
		if err != nil {
			return false, err
		}
		return strings.EqualFold(s.Kid, expect), nil
	case hintEthAddr:
		if !strings.EqualFold(s.Alg, playlist.AlgEIP191) {
			return false, nil
		}
		addr, _, err := sign.EthereumAddressFromDIDPKH(s.Kid)
		if err != nil {
			return false, nil
		}
		return strings.EqualFold(addr, hint.ethAddr), nil
	default:
		return false, fmt.Errorf("internal: unknown pubkey hint style")
	}
}

// PubkeyHint is a normalized filter for multi-signatures.
type PubkeyHint struct {
	style   hintStyle
	didKey  string
	didPKH  string
	edPub   ed25519.PublicKey
	ethAddr string // EIP-55 form
}

type hintStyle int

const (
	hintDIDKey hintStyle = iota
	hintDIDPKH
	hintEd25519Raw
	hintEthAddr
)

// ParsePubkeyHint understands:
//   - did:key:… (match kid)
//   - did:pkh:… (match kid)
//   - 64 hex chars → Ed25519 raw public key → derive did:key for matching
//   - Ethereum address (0x-prefixed 20-byte hex) → match eip191 signatures by address (any chain)
func ParsePubkeyHint(s string) (*PubkeyHint, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty --pubkey")
	}
	low := strings.ToLower(s)
	if strings.HasPrefix(low, "did:key:") {
		return &PubkeyHint{style: hintDIDKey, didKey: s}, nil
	}
	if strings.HasPrefix(low, "did:pkh:") {
		return &PubkeyHint{style: hintDIDPKH, didPKH: s}, nil
	}

	hs := strings.TrimPrefix(s, "0x")
	// Ethereum address (20 bytes) without 0x prefix
	if len(hs) == 40 && common.IsHexAddress("0x"+hs) {
		return &PubkeyHint{style: hintEthAddr, ethAddr: common.HexToAddress("0x" + hs).Hex()}, nil
	}
	if len(hs) == 64 {
		b, err := hex.DecodeString(hs)
		if err != nil {
			return nil, fmt.Errorf("decode hex public key: %w", err)
		}
		if len(b) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("Ed25519 public key must be %d bytes, got %d", ed25519.PublicKeySize, len(b))
		}
		return &PubkeyHint{style: hintEd25519Raw, edPub: ed25519.PublicKey(b)}, nil
	}

	addrStr := s
	if common.IsHexAddress(addrStr) {
		return &PubkeyHint{style: hintEthAddr, ethAddr: common.HexToAddress(addrStr).Hex()}, nil
	}
	return nil, fmt.Errorf("unrecognized --pubkey (want did:key, did:pkh, 32-byte hex, or Ethereum address)")
}

func ed25519PubFromFlag(s string) (ed25519.PublicKey, error) {
	h, err := ParsePubkeyHint(s)
	if err != nil {
		return nil, err
	}
	switch h.style {
	case hintEd25519Raw:
		return h.edPub, nil
	case hintDIDKey:
		return sign.Ed25519PublicKeyFromDIDKey(h.didKey)
	default:
		return nil, fmt.Errorf("expected 32-byte Ed25519 hex or did:key for legacy verification")
	}
}
