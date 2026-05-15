package ask

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// FieldTracker tracks completed fields for display during interactive sessions.
type FieldTracker struct {
	Fields         []CompletedField // exported for access in create package
	lastDisplayLen int              // number of fields displayed last time
}

// CompletedField represents a single completed field.
type CompletedField struct {
	Label string
	Value string
}

// NewFieldTracker creates a new field tracker for an interactive session.
func NewFieldTracker() *FieldTracker {
	return &FieldTracker{
		Fields:         make([]CompletedField, 0),
		lastDisplayLen: 0,
	}
}

// Add records a completed field with its label and value.
func (ft *FieldTracker) Add(label, value string) {
	ft.Fields = append(ft.Fields, CompletedField{
		Label: label,
		Value: value,
	})
}

// Display prints all completed fields with their values.
// Only displays new fields since last display.
func (ft *FieldTracker) Display() {
	if len(ft.Fields) == 0 {
		return
	}

	// Only display fields we haven't displayed yet
	if ft.lastDisplayLen >= len(ft.Fields) {
		return
	}

	// Display new completed fields with color
	cyan := color.New(color.FgCyan).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()
	checkmark := color.New(color.FgGreen).Sprint("✓")

	for i := ft.lastDisplayLen; i < len(ft.Fields); i++ {
		f := ft.Fields[i]
		fmt.Printf("%s %s %s\n", checkmark, cyan(f.Label+":"), white(f.Value))
	}

	ft.lastDisplayLen = len(ft.Fields)
}

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
	return ConfirmWithTracker(nil, label, defYes)
}

// ConfirmWithTracker prompts y/n with field tracking support.
func ConfirmWithTracker(tracker *FieldTracker, label string, defYes bool) (bool, error) {
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
	result := resolveConfirm(s, defYes)

	// Track the completed field
	if tracker != nil {
		value := "no"
		if result {
			value = "yes"
		}
		tracker.Add(label, value)
	}

	return result, nil
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
	return LineWithTracker(nil, label, def, allowEmpty, validate)
}

// LineWithTracker reads one line with field tracking support.
func LineWithTracker(tracker *FieldTracker, label string, def string, allowEmpty bool, validate func(string) error) (string, error) {
	gray := color.New(color.FgHiBlack).SprintFunc()

	// Build prompt label
	promptLabel := label

	// Create templates for showing default as a hint
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	if def != "" {
		// Show default as gray hint
		promptLabel = fmt.Sprintf("%s %s", label, gray(fmt.Sprintf("(default: %s)", def)))
	}

	prompt := promptui.Prompt{
		Label:     promptLabel,
		Templates: templates,
		Validate: func(raw string) error {
			return validateLinePrompt(raw, def, allowEmpty, validate)
		},
	}

	raw, err := prompt.Run()
	if err != nil {
		return "", err
	}

	result, err := finalizeLine(raw, def, allowEmpty, validate)
	if err != nil {
		return "", err
	}

	// Track the completed field (caller will update value if needed before next prompt)
	if tracker != nil {
		displayValue := result
		if displayValue == "" {
			displayValue = "(empty)"
		}
		tracker.Add(label, displayValue)
	}

	return result, nil
}

// UpdateLastField updates the value of the most recently added field.
// Use this when a value is generated after the initial prompt.
func (ft *FieldTracker) UpdateLastField(value string) {
	if len(ft.Fields) > 0 {
		ft.Fields[len(ft.Fields)-1].Value = value
	}
}

func Select(label string, items []string) (int, string, error) {
	return SelectWithTracker(nil, label, items)
}

// SelectWithTracker shows a selection prompt with field tracking support.
func SelectWithTracker(tracker *FieldTracker, label string, items []string) (int, string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}
	idx, value, err := prompt.Run()
	if err != nil {
		return idx, value, err
	}

	// Track the completed field
	if tracker != nil {
		tracker.Add(label, value)
	}

	return idx, value, nil
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
	return FloatEmptyOKWithTracker(nil, label)
}

// FloatEmptyOKWithTracker reads an optional float with field tracking support.
func FloatEmptyOKWithTracker(tracker *FieldTracker, label string) (*float64, error) {
	s, err := LineWithTracker(tracker, label+" (optional)", "", true, nil)
	if err != nil {
		return nil, err
	}
	return ParseOptionalFloat(s)
}

func IntEmptyOK(label string) (*int, error) {
	return IntEmptyOKWithTracker(nil, label)
}

// IntEmptyOKWithTracker reads an optional int with field tracking support.
func IntEmptyOKWithTracker(tracker *FieldTracker, label string) (*int, error) {
	s, err := LineWithTracker(tracker, label+" (optional)", "", true, nil)
	if err != nil {
		return nil, err
	}
	return ParseOptionalInt(s)
}
