package validateerr

import (
	"testing"

	dp1 "github.com/display-protocol/dp1-go"
)

func TestOnlyMissingSignature(t *testing.T) {
	const minimalItem = `"items":[{"source":"https://example.invalid/i.json"}]`

	cases := []struct {
		name string
		doc  string
		fn   func([]byte) error
		want bool
	}{
		{
			name: "playlist_unsigned",
			doc:  `{"dpVersion":"1.1.0","title":"x",` + minimalItem + `}`,
			fn: func(b []byte) error {
				_, err := dp1.ParseAndValidatePlaylist(b)
				return err
			},
			want: true,
		},
		{
			name: "playlist_empty_signatures",
			doc:  `{"dpVersion":"1.1.0","title":"x",` + minimalItem + `,"signatures":[]}`,
			fn: func(b []byte) error {
				_, err := dp1.ParseAndValidatePlaylist(b)
				return err
			},
			want: true,
		},
		{
			name: "playlist_unsigned_bad_title",
			doc:  `{"dpVersion":"1.1.0","title":"",` + minimalItem + `}`,
			fn: func(b []byte) error {
				_, err := dp1.ParseAndValidatePlaylist(b)
				return err
			},
			want: false,
		},
		{
			name: "playlist_invalid_signature_field",
			doc: `{"dpVersion":"1.1.0","title":"x",` + minimalItem + `,
				"signatures":[{"alg":"rsa999","kid":"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
				"ts":"2025-01-01T00:00:00Z","payload_hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"role":"curator","sig":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}]}`,
			fn: func(b []byte) error {
				_, err := dp1.ParseAndValidatePlaylist(b)
				return err
			},
			want: false,
		},
		{
			name: "group_unsigned",
			doc:  `{"id":"385f79b6-a45f-4c1c-8080-e93a192adccc","title":"g","playlists":["https://p"],"created":"2025-01-01T00:00:00Z"}`,
			fn: func(b []byte) error {
				_, err := dp1.ParseAndValidatePlaylistGroup(b)
				return err
			},
			want: true,
		},
		{
			name: "channel_unsigned",
			doc:  `{"id":"385f79b6-a45f-4c1c-8080-e93a192adccc","slug":"s","title":"c","version":"1.0.0","created":"2025-01-01T00:00:00Z","playlists":["https://p"]}`,
			fn: func(b []byte) error {
				_, err := dp1.ParseAndValidateChannel(b)
				return err
			},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn([]byte(tc.doc))
			if err == nil {
				t.Fatal("expected validation error")
			}
			got := OnlyMissingSignature(err)
			if got != tc.want {
				t.Fatalf("OnlyMissingSignature() = %v, want %v; err=%v", got, tc.want, err)
			}
		})
	}
}

func TestOnlyMissingSignature_nilAndNonValidation(t *testing.T) {
	if OnlyMissingSignature(nil) {
		t.Fatal("nil error should be false")
	}
	if OnlyMissingSignature(dp1.ErrValidation) {
		t.Fatal("bare ErrValidation without jsonschema detail should be false")
	}
}
