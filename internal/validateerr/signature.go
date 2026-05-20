// Package validateerr classifies dp1-go JSON Schema validation failures for CLI policy
// (e.g. tolerating unsigned drafts when only signature requirements fail).
package validateerr

import (
	"errors"
	"strings"

	dp1 "github.com/display-protocol/dp1-go"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

// OnlyMissingSignature reports whether err is a JSON Schema validation failure and every
// leaf cause is solely about absent or empty top-level signature material (v1.1+ signatures
// array or legacy signature field). Invalid signature content at signatures/0/... or a
// malformed legacy signature value is not treated as missing-only.
func OnlyMissingSignature(err error) bool {
	if err == nil || !errors.Is(err, dp1.ErrValidation) {
		return false
	}
	var ve *jsonschema.ValidationError
	if !errors.As(err, &ve) {
		return false
	}
	leaves := collectLeaves(ve)
	if len(leaves) == 0 {
		return false
	}
	for _, leaf := range leaves {
		if !isMissingSignatureLeaf(leaf) {
			return false
		}
	}
	return true
}

func collectLeaves(v *jsonschema.ValidationError) []*jsonschema.ValidationError {
	if v == nil {
		return nil
	}
	if len(v.Causes) == 0 {
		return []*jsonschema.ValidationError{v}
	}
	var out []*jsonschema.ValidationError
	for _, c := range v.Causes {
		out = append(out, collectLeaves(c)...)
	}
	return out
}

func isMissingSignatureLeaf(v *jsonschema.ValidationError) bool {
	loc := v.InstanceLocation
	msg := v.Error()
	switch len(loc) {
	case 0:
		return strings.Contains(msg, "missing property 'signatures'") ||
			strings.Contains(msg, "missing property 'signature'")
	case 1:
		if loc[0] == "signatures" && strings.Contains(msg, "minItems") {
			return true
		}
	}
	return false
}
