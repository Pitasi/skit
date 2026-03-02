package internal_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pitasi/skit/internal/parser"
	"github.com/Pitasi/skit/internal/site"
)

func TestBuildEndToEnd(t *testing.T) {
	// Create a temp directory with a deck file.
	tmpDir := t.TempDir()
	deckContent := `---
title: Integration Test
author: Tester
aspectRatio: "16:9"
---

# Hello World

These are speaker notes.

	This is visible on the slide.

---

# Slide Two

:::slide
Directive content.
:::

Notes for slide two.
`
	deckPath := filepath.Join(tmpDir, "deck.md")
	if err := os.WriteFile(deckPath, []byte(deckContent), 0o644); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(tmpDir, "dist")

	// Parse.
	deck, err := parser.ParseFile(deckPath, parser.Options{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(deck.Slides) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(deck.Slides))
	}

	// Build site (includes markdown rendering).
	if err := site.Build(deck, site.BuildOptions{
		InputFile:   deckPath,
		OutputDir:   outDir,
		BaseURL:     "/",
		AspectRatio: "16:9",
	}); err != nil {
		t.Fatalf("build: %v", err)
	}

	// Verify outputs exist.
	assertFileExists(t, filepath.Join(outDir, "index.html"))
	assertFileExists(t, filepath.Join(outDir, "assets", "reveal", "dist", "reveal.js"))
	assertFileExists(t, filepath.Join(outDir, "assets", "reveal", "dist", "reveal.css"))
	assertFileExists(t, filepath.Join(outDir, "assets", "reveal", "plugin", "notes", "notes.js"))
	assertFileExists(t, filepath.Join(outDir, "assets", "theme.css"))

	// Verify HTML content.
	htmlBytes, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(htmlBytes)

	if !strings.Contains(html, "<title>Integration Test</title>") {
		t.Error("missing title")
	}
	if !strings.Contains(html, "Hello World") {
		t.Error("missing slide 1 heading")
	}
	if !strings.Contains(html, "This is visible on the slide") {
		t.Error("missing promoted content")
	}
	if !strings.Contains(html, "Directive content") {
		t.Error("missing directive content")
	}
	if !strings.Contains(html, `class="notes"`) {
		t.Error("missing notes aside")
	}
}

func TestBuildEndToEnd_ThemeAndTransition(t *testing.T) {
	tmpDir := t.TempDir()
	deckContent := `---
title: Styled Deck
theme: moon
transition: fade
---

# Slide One

Notes here.

	Visible content.
`
	deckPath := filepath.Join(tmpDir, "deck.md")
	if err := os.WriteFile(deckPath, []byte(deckContent), 0o644); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(tmpDir, "dist")

	deck, err := parser.ParseFile(deckPath, parser.Options{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if err := site.Build(deck, site.BuildOptions{
		InputFile:  deckPath,
		OutputDir:  outDir,
		BaseURL:    "/",
		Theme:      "moon",
		Transition: "fade",
	}); err != nil {
		t.Fatalf("build: %v", err)
	}

	htmlBytes, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(htmlBytes)

	if !strings.Contains(html, "transition: 'fade'") {
		t.Error("expected transition: 'fade' in Reveal.initialize()")
	}

	// Verify theme.css was written (moon built-in).
	assertFileExists(t, filepath.Join(outDir, "assets", "theme.css"))
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %s", path)
	}
}
