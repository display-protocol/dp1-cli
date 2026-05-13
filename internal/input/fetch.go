package input

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const userAgent = "dp1-cli/1.0"

// ReadSource loads DP-1 JSON from stdin ("-" or empty), HTTP(S) URL, a file path, or a base64 string.
func ReadSource(source string) ([]byte, error) {
	src := strings.TrimSpace(source)
	if src == "" || src == "-" {
		return io.ReadAll(os.Stdin)
	}

	if u, err := url.Parse(src); err == nil && u.Scheme != "" && u.Host != "" {
		switch strings.ToLower(u.Scheme) {
		case "http", "https":
			return fetchURL(src)
		default:
			return nil, fmt.Errorf("unsupported URL scheme %q (use http or https)", u.Scheme)
		}
	}

	if data, err := os.ReadFile(src); err == nil {
		return data, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// Inline base64 payload
	if dec, err := base64.StdEncoding.DecodeString(strings.TrimSpace(src)); err == nil && len(dec) > 0 {
		return dec, nil
	}

	return nil, fmt.Errorf("cannot read source %q: not a valid file, URL, or base64 payload", source)
}

func fetchURL(rawURL string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
