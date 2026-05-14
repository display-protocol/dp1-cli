package create

import "testing"

func TestNonEmptyValidator(t *testing.T) {
	t.Parallel()
	v := nonEmpty()
	if err := v(""); err == nil {
		t.Fatal("empty: expected error")
	}
	if err := v("  "); err == nil {
		t.Fatal("whitespace: expected error")
	}
	if err := v("ok"); err != nil {
		t.Fatal(err)
	}
}
