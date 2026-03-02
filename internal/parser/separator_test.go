package parser

import (
	"strings"
	"testing"
)

func TestSeparateContent_TabPromotion(t *testing.T) {
	chunk := "# Title\n\nnotes paragraph\n\n\tslide paragraph\n"
	slide := SeparateContent(chunk, 0)

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
	slide := SeparateContent(chunk, 0)

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
	slide := SeparateContent(chunk, 0)

	if strings.Contains(slide.SlideMarkdown, "private comment") {
		t.Errorf("comment should not appear in slide markdown")
	}
	if strings.Contains(slide.NotesMarkdown, "private comment") {
		t.Errorf("comment should not appear in notes markdown")
	}
}

func TestSeparateContent_FirstHeadingIsSlideVisible(t *testing.T) {
	chunk := "# My Title\n\nsome notes\n"
	slide := SeparateContent(chunk, 0)

	if slide.Title != "My Title" {
		t.Errorf("expected title 'My Title', got %q", slide.Title)
	}
	if !strings.Contains(slide.SlideMarkdown, "# My Title") {
		t.Errorf("heading should be in slide markdown")
	}
}

func TestSeparateContent_MediaRefs(t *testing.T) {
	chunk := "# Title\n\n\t![photo](images/photo.png)\n\nnotes with ![diagram](https://example.com/d.png)\n"
	slide := SeparateContent(chunk, 0)

	if len(slide.MediaRefs) != 1 {
		t.Fatalf("expected 1 local media ref, got %d", len(slide.MediaRefs))
	}
	if slide.MediaRefs[0] != "images/photo.png" {
		t.Errorf("expected 'images/photo.png', got %q", slide.MediaRefs[0])
	}
}

func TestSeparateContent_AllNotesDefault(t *testing.T) {
	chunk := "Just a paragraph with no heading.\n\nAnother paragraph.\n"
	slide := SeparateContent(chunk, 0)

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
