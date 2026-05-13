package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func marshallPretty(v any) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte{'\n'}), nil
}

func writeJSONDocument(dest string, v any, perm os.FileMode) error {
	raw, err := marshallPretty(v)
	if err != nil {
		return err
	}
	line := append(append([]byte(nil), raw...), '\n')

	if dest == "" || dest == "-" {
		_, err = os.Stdout.Write(line)
		return err
	}
	if err := os.WriteFile(dest, line, perm); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	return nil
}

func writeRawDocument(cmd *cobra.Command, dest string, data []byte) error {
	out := cmd.OutOrStdout()
	if dest != "" && dest != "-" {
		return os.WriteFile(dest, data, 0o644)
	}
	_, err := out.Write(data)
	return err
}
