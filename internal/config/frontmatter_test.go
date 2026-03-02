package config

import (
	"strings"
	"testing"
)

func TestParseFrontMatter_Basic(t *testing.T) {
	content := "---\ntitle: Test\nauthor: Me\n---\n# Slide 1\n"
	meta, body, err := ParseFrontMatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", meta.Title)
	}
	if meta.Author != "Me" {
		t.Errorf("expected author 'Me', got %q", meta.Author)
	}
	if !strings.Contains(body, "# Slide 1") {
		t.Errorf("expected body to contain slide content, got %q", body)
	}
}

func TestParseFrontMatter_NoFrontMatter(t *testing.T) {
	content := "# Just a slide\n"
	meta, body, err := ParseFrontMatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Title != "" {
		t.Errorf("expected empty title, got %q", meta.Title)
	}
	if body != content {
		t.Errorf("expected full content as body")
	}
}

func TestParseFrontMatter_WithBOM(t *testing.T) {
	content := "\xef\xbb\xbf---\ntitle: BOM Test\n---\nbody\n"
	meta, _, err := ParseFrontMatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Title != "BOM Test" {
		t.Errorf("expected title 'BOM Test', got %q", meta.Title)
	}
}

func TestParseFrontMatter_DotsClosing(t *testing.T) {
	content := "---\ntitle: Dots\n...\nbody\n"
	meta, body, err := ParseFrontMatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Title != "Dots" {
		t.Errorf("expected title 'Dots', got %q", meta.Title)
	}
	if !strings.Contains(body, "body") {
		t.Errorf("expected body content")
	}
}

func TestParseFrontMatter_Unclosed(t *testing.T) {
	content := "---\ntitle: Unclosed\n"
	_, _, err := ParseFrontMatter(content)
	if err == nil {
		t.Fatal("expected error for unclosed front matter")
	}
}
