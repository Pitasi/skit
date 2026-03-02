package parser

import (
	"strings"
	"testing"
)

func TestParse_UnclosedDirectiveReturnsError(t *testing.T) {
	content := "---\ntitle: Test\n---\n\n# Slide\n\n:::slide\ncontent without closing\n"
	_, err := Parse(content, Options{})
	if err == nil {
		t.Fatal("expected error for unclosed :::slide block")
	}
	if !strings.Contains(err.Error(), "unclosed :::slide block") {
		t.Errorf("expected error about unclosed directive, got: %v", err)
	}
}

func TestParse_ClosedDirectiveNoError(t *testing.T) {
	content := "---\ntitle: Test\n---\n\n# Slide\n\n:::slide\ncontent\n:::\n"
	deck, err := Parse(content, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deck.Slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(deck.Slides))
	}
}
