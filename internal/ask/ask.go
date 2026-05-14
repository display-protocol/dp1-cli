package ask

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

// validateConfirmInput is the prompt validator for [Confirm].
func validateConfirmInput(s string) error {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return nil
	}
	switch s {
	case "y", "yes", "n", "no":
		return nil
	default:
		return fmt.Errorf(`type "y" or "n"`)
	}
}

// resolveConfirm maps trimmed lower-case user input to a bool; empty means defYes.
func resolveConfirm(trimmedLower string, defYes bool) bool {
	if trimmedLower == "" {
		return defYes
	}
	return trimmedLower == "y" || trimmedLower == "yes"
}

// Confirm prompts y/n ; empty uses defYes.
func Confirm(label string, defYes bool) (bool, error) {
	opts := "(y/N)"
	if defYes {
		opts = "(Y/n)"
	}
	prompt := promptui.Prompt{
		Label:    label + " " + opts,
		Validate: validateConfirmInput,
	}
	s, err := prompt.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrInterrupt) {
			return false, fmt.Errorf("%w", err)
		}
		return false, err
	}
	s = strings.TrimSpace(strings.ToLower(s))
	return resolveConfirm(s, defYes), nil
}

// normalizeLineValue trims and applies default when the user left the line empty.
func normalizeLineValue(raw string, def string) string {
	s := strings.TrimSpace(raw)
	if s == "" && def != "" {
		s = strings.TrimSpace(def)
	}
	return s
}

// validateLinePrompt is the prompt-time validator for [Line] (matches promptui input before return).
func validateLinePrompt(raw string, def string, allowEmpty bool, validate func(string) error) error {
	got := strings.TrimSpace(raw)
	if got == "" && def != "" {
		got = strings.TrimSpace(def)
	}
	if !allowEmpty && strings.TrimSpace(raw) == "" && def == "" {
		return fmt.Errorf("required")
	}
	if !allowEmpty && got == "" {
		return fmt.Errorf("required")
	}
	if validate != nil {
		chk := strings.TrimSpace(raw)
		if chk == "" && def != "" {
			chk = strings.TrimSpace(def)
		}
		if chk == "" && allowEmpty {
			return nil
		}
		return validate(chk)
	}
	return nil
}

// finalizeLine applies default trimming and optional validator after a successful read.
func finalizeLine(raw string, def string, allowEmpty bool, validate func(string) error) (string, error) {
	s := normalizeLineValue(raw, def)
	if validate != nil {
		chk := s
		if chk == "" && allowEmpty {
			return "", nil
		}
		if err := validate(chk); err != nil {
			return "", err
		}
	}
	return s, nil
}

// Line reads one line with optional validation. Trimmed spaces; empty substitutes def when helpful.
func Line(label string, def string, allowEmpty bool, validate func(string) error) (string, error) {
	show := label
	if def != "" && !strings.Contains(strings.ToLower(label), "optional") {
		show = fmt.Sprintf("%s [%s]", label, def)
	}
	prompt := promptui.Prompt{
		Label: show,
		Validate: func(raw string) error {
			return validateLinePrompt(raw, def, allowEmpty, validate)
		},
	}
	raw, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return finalizeLine(raw, def, allowEmpty, validate)
}

func Select(label string, items []string) (int, string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}
	return prompt.Run()
}

// ParseOptionalFloat returns nil for blank input; otherwise parses a float64.
func ParseOptionalFloat(s string) (*float64, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number")
	}
	return &v, nil
}

// ParseOptionalInt returns nil for blank input; otherwise parses an int.
func ParseOptionalInt(s string) (*int, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("invalid integer")
	}
	return &v, nil
}

func FloatEmptyOK(label string) (*float64, error) {
	s, err := Line(label+" (optional)", "", true, nil)
	if err != nil {
		return nil, err
	}
	return ParseOptionalFloat(s)
}

func IntEmptyOK(label string) (*int, error) {
	s, err := Line(label+" (optional)", "", true, nil)
	if err != nil {
		return nil, err
	}
	return ParseOptionalInt(s)
}
