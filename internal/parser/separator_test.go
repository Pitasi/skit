package parser

import (
	"strings"
	"testing"
)

// helper that calls SeparateContent and fails the test on error.
func mustSeparate(t *testing.T, chunk string, index int) (slide struct {
	Title         string
	Layout        string
	SlideMarkdown string
	NotesMarkdown string
	MediaRefs     []string
	CellCount     int
	CellMarkdowns []string
}) {
	t.Helper()
	s, err := SeparateContent(chunk, index)
	if err != nil {
		t.Fatalf("SeparateContent returned unexpected error: %v", err)
	}
	slide.Title = s.Title
	slide.Layout = s.Layout
	slide.SlideMarkdown = s.SlideMarkdown
	slide.NotesMarkdown = s.NotesMarkdown
	slide.MediaRefs = s.MediaRefs
	slide.CellCount = len(s.Cells)
	for _, c := range s.Cells {
		slide.CellMarkdowns = append(slide.CellMarkdowns, c.Markdown)
	}
	return slide
}

func TestSeparateContent_TabPromotion(t *testing.T) {
	chunk := "# Title\n\nnotes paragraph\n\n\tslide paragraph\n"
	slide := mustSeparate(t, chunk, 0)

	if slide.Title != "Title" {
		t.Errorf("expected title 'Title', got %q", slide.Title)
	}
	if !strings.Contains(slide.SlideMarkdown, "slide paragraph") {
		t.Errorf("expected slide markdown to contain promoted text, got %q", slide.SlideMarkdown)
	}
	if !strings.Contains(slide.NotesMarkdown, "notes paragraph") {
		t.Errorf("expected notes markdown to contain notes text, got %q", slide.NotesMarkdown)
	}
	// Tab should be stripped from promoted content.
	if strings.Contains(slide.SlideMarkdown, "\t") {
		t.Errorf("tab should be stripped from promoted content")
	}
}

func TestSeparateContent_DirectiveBlock(t *testing.T) {
	chunk := "# Title\n\nnotes here\n\n:::slide\nvisible on slide\n:::\n"
	slide := mustSeparate(t, chunk, 0)

	if !strings.Contains(slide.SlideMarkdown, "visible on slide") {
		t.Errorf("expected directive content in slide markdown, got %q", slide.SlideMarkdown)
	}
	if strings.Contains(slide.SlideMarkdown, ":::") {
		t.Errorf("directive markers should be stripped")
	}
	if strings.Contains(slide.NotesMarkdown, "visible on slide") {
		t.Errorf("directive content should not appear in notes")
	}
}

func TestSeparateContent_CommentStripping(t *testing.T) {
	chunk := "# Title\n\n// private comment\n\nnotes here\n"
	slide := mustSeparate(t, chunk, 0)

	if strings.Contains(slide.SlideMarkdown, "private comment") {
		t.Errorf("comment should not appear in slide markdown")
	}
	if strings.Contains(slide.NotesMarkdown, "private comment") {
		t.Errorf("comment should not appear in notes markdown")
	}
}

func TestSeparateContent_FirstHeadingIsSlideVisible(t *testing.T) {
	chunk := "# My Title\n\nsome notes\n"
	slide := mustSeparate(t, chunk, 0)

	if slide.Title != "My Title" {
		t.Errorf("expected title 'My Title', got %q", slide.Title)
	}
	if !strings.Contains(slide.SlideMarkdown, "# My Title") {
		t.Errorf("heading should be in slide markdown")
	}
}

func TestSeparateContent_MediaRefs(t *testing.T) {
	chunk := "# Title\n\n\t![photo](images/photo.png)\n\nnotes with ![diagram](https://example.com/d.png)\n"
	slide := mustSeparate(t, chunk, 0)

	if len(slide.MediaRefs) != 1 {
		t.Fatalf("expected 1 local media ref, got %d", len(slide.MediaRefs))
	}
	if slide.MediaRefs[0] != "images/photo.png" {
		t.Errorf("expected 'images/photo.png', got %q", slide.MediaRefs[0])
	}
}

func TestSeparateContent_AllNotesDefault(t *testing.T) {
	chunk := "Just a paragraph with no heading.\n\nAnother paragraph.\n"
	slide := mustSeparate(t, chunk, 0)

	if slide.Title != "" {
		t.Errorf("expected empty title, got %q", slide.Title)
	}
	if slide.SlideMarkdown != "" {
		t.Errorf("expected empty slide markdown, got %q", slide.SlideMarkdown)
	}
	if !strings.Contains(slide.NotesMarkdown, "Just a paragraph") {
		t.Errorf("expected all content in notes")
	}
}

func TestSeparateContent_LayoutDirective(t *testing.T) {
	chunk := "# Title\n\n:::layout split\n\n\tLeft content\n\n\tRight content\n"
	slide := mustSeparate(t, chunk, 0)

	if slide.Layout != "split" {
		t.Errorf("expected layout 'split', got %q", slide.Layout)
	}
	// Layout directive should not appear in slide or notes markdown.
	if strings.Contains(slide.SlideMarkdown, ":::layout") {
		t.Errorf("layout directive should be stripped from slide markdown")
	}
	if strings.Contains(slide.NotesMarkdown, ":::layout") {
		t.Errorf("layout directive should be stripped from notes markdown")
	}
}

func TestSeparateContent_LayoutInvalidName(t *testing.T) {
	chunk := "# Title\n\n:::layout bogus\n\n\tsome content\n"
	_, err := SeparateContent(chunk, 0)
	if err == nil {
		t.Fatal("expected error for unknown layout")
	}
	if !strings.Contains(err.Error(), "unknown layout") {
		t.Errorf("expected 'unknown layout' error, got: %v", err)
	}
}

func TestSeparateContent_CellSplitting(t *testing.T) {
	chunk := "# Title\n\n\tFirst cell line 1\n\tFirst cell line 2\n\n\tSecond cell\n"
	slide := mustSeparate(t, chunk, 0)

	// Heading is one cell, then two tab-promoted blocks separated by blank line.
	if slide.CellCount != 3 {
		t.Fatalf("expected 3 cells, got %d: %v", slide.CellCount, slide.CellMarkdowns)
	}
	if !strings.Contains(slide.CellMarkdowns[0], "# Title") {
		t.Errorf("cell 0 should contain heading, got %q", slide.CellMarkdowns[0])
	}
	if !strings.Contains(slide.CellMarkdowns[1], "First cell line 1") {
		t.Errorf("cell 1 should contain first promoted block, got %q", slide.CellMarkdowns[1])
	}
	if !strings.Contains(slide.CellMarkdowns[2], "Second cell") {
		t.Errorf("cell 2 should contain second promoted block, got %q", slide.CellMarkdowns[2])
	}
}

func TestSeparateContent_NoCellsWhenNoSlideContent(t *testing.T) {
	chunk := "Just notes, nothing on slide.\n"
	slide := mustSeparate(t, chunk, 0)

	if slide.CellCount != 0 {
		t.Errorf("expected 0 cells for notes-only slide, got %d", slide.CellCount)
	}
}

func TestSeparateContent_HeadingAndTabBlockAreSeparateCells(t *testing.T) {
	chunk := "# Title\n\n\tLine one\n\tLine two\n"
	slide := mustSeparate(t, chunk, 0)

	// Heading and tab-promoted block are distinct content sources,
	// so they become separate cells even without an explicit blank line
	// in the original markdown between them.
	if slide.CellCount != 2 {
		t.Errorf("expected 2 cells, got %d: %v",
			slide.CellCount, slide.CellMarkdowns)
	}
}


