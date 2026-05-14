package cmd

import "testing"

func TestDraftPathLabel(t *testing.T) {
	if draftPathLabel("") != "-" || draftPathLabel("-") != "-" || draftPathLabel("/tmp/x.json") != "/tmp/x.json" {
		t.Fatalf("unexpected draftPathLabel mapping")
	}
}

func TestModeLabel(t *testing.T) {
	if modeLabel(false) != "all" || modeLabel(true) != "pubkey" {
		t.Fatal(modeLabel(false), modeLabel(true))
	}
}
