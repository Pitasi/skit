package render

import (
	"os"
	"strings"
	"testing"

	"github.com/Pitasi/skit/internal/parser"
)

func TestRenderDeck_Golden(t *testing.T) {
	deck, err := parser.ParseFile("../testdata/basic.md", parser.Options{})
	if err != nil {
		t.Fatal(err)
	}

	if err := RenderDeck(deck); err != nil {
		t.Fatal(err)
	}

	if len(deck.Slides) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(deck.Slides))
	}

	// Slide 0: heading + tab-promoted content.
	s0 := deck.Slides[0]
	if !strings.Contains(s0.SlideHTML, "First Slide") {
		t.Errorf("slide 0: expected heading in slide HTML")
	}
	if !strings.Contains(s0.SlideHTML, "Visible text on the slide") {
		t.Errorf("slide 0: expected promoted text in slide HTML")
	}
	if !strings.Contains(s0.NotesHTML, "Speaker notes") {
		t.Errorf("slide 0: expected notes in notes HTML")
	}

	// Slide 1: directive content.
	s1 := deck.Slides[1]
	if !strings.Contains(s1.SlideHTML, "Directive content") {
		t.Errorf("slide 1: expected directive content in slide HTML")
	}
	if !strings.Contains(s1.NotesHTML, "More speaker notes") {
		t.Errorf("slide 1: expected notes in notes HTML")
	}

	// Generate full HTML and verify structure.
	data := NewTemplateData(deck.Meta, deck.Slides, "/", "16:9", "")
	html, err := RenderHTML(data)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(html, "<title>Basic Test</title>") {
		t.Error("expected title in HTML")
	}
	if !strings.Contains(html, `content="Test Author"`) {
		t.Error("expected author meta in HTML")
	}
	if !strings.Contains(html, "reveal.js") {
		t.Error("expected reveal.js script reference")
	}
	if strings.Count(html, "<section") != 2 {
		t.Errorf("expected 2 sections, got %d", strings.Count(html, "<section"))
	}

	// Update golden file if UPDATE_GOLDEN is set.
	if os.Getenv("UPDATE_GOLDEN") != "" {
		os.WriteFile("../testdata/basic.golden.html", []byte(html), 0o644)
	}

	// If golden file exists, compare.
	golden, err := os.ReadFile("../testdata/basic.golden.html")
	if err == nil {
		if string(golden) != html {
			t.Error("HTML output differs from golden file. Set UPDATE_GOLDEN=1 to update.")
		}
	}
}
