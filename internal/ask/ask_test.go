package ask

import (
	"errors"
	"fmt"
	"testing"
)

func TestValidateConfirmInput(t *testing.T) {
	if err := validateConfirmInput(""); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"y", "yes", "n", "no", "  Y  "} {
		if err := validateConfirmInput(s); err != nil {
			t.Fatalf("%q: %v", s, err)
		}
	}
	if err := validateConfirmInput("maybe"); err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveConfirm(t *testing.T) {
	if !resolveConfirm("", true) || resolveConfirm("", false) {
		t.Fatal("empty defYes boundary")
	}
	if !resolveConfirm("yes", false) || !resolveConfirm("y", false) {
		t.Fatal("yes")
	}
	if resolveConfirm("no", true) || resolveConfirm("n", true) {
		t.Fatal("no")
	}
}

func TestNormalizeLineValue(t *testing.T) {
	if got := normalizeLineValue("  ", "def"); got != "def" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeLineValue(" x ", ""); got != "x" {
		t.Fatalf("got %q", got)
	}
}

func TestValidateLinePrompt(t *testing.T) {
	vErr := fmt.Errorf("boom")
	validate := func(s string) error {
		if s == "bad" {
			return vErr
		}
		return nil
	}

	if err := validateLinePrompt("", "", false, nil); err == nil {
		t.Fatal("required empty")
	}
	if err := validateLinePrompt("", "def", false, validate); err != nil {
		t.Fatal(err)
	}
	if err := validateLinePrompt("", "", true, validate); err != nil {
		t.Fatal(err)
	}
	if err := validateLinePrompt("bad", "", false, validate); !errors.Is(err, vErr) {
		t.Fatalf("got %v", err)
	}
}

func TestFinalizeLine(t *testing.T) {
	got, err := finalizeLine("", "x", false, func(s string) error {
		if s != "x" {
			return fmt.Errorf("want x")
		}
		return nil
	})
	if err != nil || got != "x" {
		t.Fatalf("%q %v", got, err)
	}

	got, err = finalizeLine("", "", true, func(string) error { return fmt.Errorf("no") })
	if err != nil || got != "" {
		t.Fatalf("allow empty: %q %v", got, err)
	}

	_, err = finalizeLine("bad", "", false, func(s string) error {
		return fmt.Errorf("invalid")
	})
	if err == nil {
		t.Fatal("expected validator error")
	}
}

func TestParseOptionalFloat(t *testing.T) {
	v, err := ParseOptionalFloat("")
	if err != nil || v != nil {
		t.Fatalf("%v %v", v, err)
	}
	v, err = ParseOptionalFloat("  3.5 ")
	if err != nil || v == nil || *v != 3.5 {
		t.Fatalf("%v %v", v, err)
	}
	if _, err := ParseOptionalFloat("x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseOptionalInt(t *testing.T) {
	v, err := ParseOptionalInt("")
	if err != nil || v != nil {
		t.Fatalf("%v %v", v, err)
	}
	v, err = ParseOptionalInt(" -7 ")
	if err != nil || v == nil || *v != -7 {
		t.Fatalf("%v %v", v, err)
	}
	if _, err := ParseOptionalInt("3.14"); err == nil {
		t.Fatal("expected error")
	}
}
