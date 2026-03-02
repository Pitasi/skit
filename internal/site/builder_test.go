package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pitasi/skit/internal/model"
)

func TestRewriteMediaPathsInMarkdown_OnlyRewritesImageSyntax(t *testing.T) {
	deck := &model.Deck{
		Slides: []model.Slide{
			{
				Index:         0,
				SlideMarkdown: "![photo](logo.png)\n\nSee logo.png for details.",
				NotesMarkdown: "The file logo.png is referenced as ![alt](logo.png).",
				MediaRefs:     []string{"logo.png"},
			},
		},
	}

	rewriteMediaPathsInMarkdown(deck, "/")

	s := deck.Slides[0]

	// Image syntax should be rewritten.
	if !strings.Contains(s.SlideMarkdown, "![photo](/media/logo.png)") {
		t.Errorf("expected image src rewritten in slide markdown, got: %s", s.SlideMarkdown)
	}
	// Prose text should NOT be rewritten.
	if !strings.Contains(s.SlideMarkdown, "See logo.png for details") {
		t.Errorf("expected prose text unchanged in slide markdown, got: %s", s.SlideMarkdown)
	}

	// Notes: image rewritten, prose unchanged.
	if !strings.Contains(s.NotesMarkdown, "![alt](/media/logo.png)") {
		t.Errorf("expected image src rewritten in notes markdown, got: %s", s.NotesMarkdown)
	}
	if !strings.Contains(s.NotesMarkdown, "The file logo.png is referenced") {
		t.Errorf("expected prose text unchanged in notes markdown, got: %s", s.NotesMarkdown)
	}
}

func TestRewriteMediaPathsInMarkdown_WithBaseURL(t *testing.T) {
	deck := &model.Deck{
		Slides: []model.Slide{
			{
				Index:         0,
				SlideMarkdown: "![img](images/photo.png)",
				MediaRefs:     []string{"images/photo.png"},
			},
		},
	}

	rewriteMediaPathsInMarkdown(deck, "/my-repo")

	s := deck.Slides[0]
	if !strings.Contains(s.SlideMarkdown, "![img](/my-repo/media/images/photo.png)") {
		t.Errorf("expected base URL in rewritten path, got: %s", s.SlideMarkdown)
	}
}

func TestRewriteMediaPathsInMarkdown_SkipsURLs(t *testing.T) {
	deck := &model.Deck{
		Slides: []model.Slide{
			{
				Index:         0,
				SlideMarkdown: "![ext](https://example.com/img.png)",
				MediaRefs:     []string{}, // no local refs
			},
		},
	}

	rewriteMediaPathsInMarkdown(deck, "/")

	s := deck.Slides[0]
	if !strings.Contains(s.SlideMarkdown, "https://example.com/img.png") {
		t.Errorf("expected external URL unchanged, got: %s", s.SlideMarkdown)
	}
}

func TestCopyMedia_RejectsPathTraversal(t *testing.T) {
	inputDir := t.TempDir()
	outDir := t.TempDir()

	deck := &model.Deck{
		Slides: []model.Slide{
			{
				Index:     0,
				MediaRefs: []string{"../../etc/passwd"},
			},
		},
	}

	err := copyMedia(deck, inputDir, outDir)
	if err == nil {
		t.Fatal("expected error for path traversal media ref")
	}
	if !strings.Contains(err.Error(), "resolves outside input directory") {
		t.Errorf("expected path traversal error, got: %v", err)
	}

	// Verify nothing was written to the output media directory.
	mediaDir := filepath.Join(outDir, "media")
	if _, statErr := os.Stat(mediaDir); statErr == nil {
		entries, _ := os.ReadDir(mediaDir)
		if len(entries) > 0 {
			t.Error("expected no files in media dir after rejected traversal")
		}
	}
}

func TestCopyMedia_AcceptsValidRef(t *testing.T) {
	inputDir := t.TempDir()
	outDir := t.TempDir()

	// Create a valid media file.
	imgDir := filepath.Join(inputDir, "images")
	if err := os.MkdirAll(imgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imgDir, "photo.png"), []byte("fake png"), 0o644); err != nil {
		t.Fatal(err)
	}

	deck := &model.Deck{
		Slides: []model.Slide{
			{
				Index:     0,
				MediaRefs: []string{"images/photo.png"},
			},
		},
	}

	if err := copyMedia(deck, inputDir, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the file was copied.
	copied := filepath.Join(outDir, "media", "images", "photo.png")
	if _, err := os.Stat(copied); err != nil {
		t.Errorf("expected copied file at %s", copied)
	}
}

