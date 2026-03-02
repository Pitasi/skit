package parser

import (
	"testing"
)

func TestSplitSlides_Basic(t *testing.T) {
	body := "slide 1\n---\nslide 2\n---\nslide 3\n"
	slides := SplitSlides(body)
	if len(slides) != 3 {
		t.Fatalf("expected 3 slides, got %d", len(slides))
	}
}

func TestSplitSlides_IgnoreCodeFence(t *testing.T) {
	body := "slide 1\n```\n---\n```\n---\nslide 2\n"
	slides := SplitSlides(body)
	if len(slides) != 2 {
		t.Fatalf("expected 2 slides (--- inside code fence ignored), got %d", len(slides))
	}
}

func TestSplitSlides_TildeFence(t *testing.T) {
	body := "slide 1\n~~~\n---\n~~~\n---\nslide 2\n"
	slides := SplitSlides(body)
	if len(slides) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(slides))
	}
}

func TestSplitSlides_SingleSlide(t *testing.T) {
	body := "just one slide\n"
	slides := SplitSlides(body)
	if len(slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(slides))
	}
}

func TestSplitSlides_EmptyBody(t *testing.T) {
	slides := SplitSlides("")
	if len(slides) != 0 {
		t.Fatalf("expected 0 slides for empty body, got %d", len(slides))
	}
}

func TestSplitByHeadings(t *testing.T) {
	chunks := []string{"# Title\nparagraph\n## Subtitle\nmore text\n"}
	result := SplitByHeadings(chunks)
	if len(result) != 2 {
		t.Fatalf("expected 2 slides from heading split, got %d", len(result))
	}
}

func TestSplitByHeadings_IgnoreCodeFence(t *testing.T) {
	// A heading inside a code fence should not trigger a split.
	chunk := "# Real Title\nsome text\n```bash\n# this is a comment\n## another comment\n```\nmore text\n"
	chunks := []string{chunk}
	result := SplitByHeadings(chunks)
	if len(result) != 1 {
		t.Fatalf("expected 1 slide (headings inside code fence ignored), got %d", len(result))
	}
}

func TestSplitByHeadings_HeadingAfterFence(t *testing.T) {
	// A heading after a code fence should still trigger a split.
	chunk := "# First\ntext\n```\ncode\n```\n## Second\nmore text\n"
	chunks := []string{chunk}
	result := SplitByHeadings(chunks)
	if len(result) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(result))
	}
}
