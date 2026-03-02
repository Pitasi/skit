package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pitasi/skit/internal/assets"
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

func TestResolveTheme_BuiltinName(t *testing.T) {
	outDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := resolveTheme(outDir, "moon"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dest := filepath.Join(outDir, "assets", "theme.css")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("theme.css not written: %v", err)
	}

	want, _ := assets.RevealFS.ReadFile("reveal/dist/theme/moon.css")
	if string(got) != string(want) {
		t.Error("theme.css content does not match embedded moon.css")
	}
}

func TestResolveTheme_CSSFilePath(t *testing.T) {
	outDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}

	customCSS := filepath.Join(t.TempDir(), "brand.css")
	if err := os.WriteFile(customCSS, []byte("/* brand */"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := resolveTheme(outDir, customCSS); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(outDir, "assets", "theme.css"))
	if string(got) != "/* brand */" {
		t.Errorf("expected custom CSS content, got: %s", got)
	}
}

func TestResolveTheme_Directory(t *testing.T) {
	outDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}

	themeDir := filepath.Join(t.TempDir(), "mytheme")
	if err := os.MkdirAll(themeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themeDir, "theme.css"), []byte("/* dir theme */"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := resolveTheme(outDir, themeDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(outDir, "assets", "theme.css"))
	if string(got) != "/* dir theme */" {
		t.Errorf("expected directory theme content, got: %s", got)
	}
}

func TestResolveTheme_DefaultWhenEmpty(t *testing.T) {
	outDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := resolveTheme(outDir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(outDir, "assets", "theme.css"))
	want, _ := assets.RevealFS.ReadFile("reveal/dist/theme/white.css")
	if string(got) != string(want) {
		t.Error("default theme should be white.css")
	}
}

func TestResolveTheme_UnknownName(t *testing.T) {
	outDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := resolveTheme(outDir, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown theme")
	}
	if !strings.Contains(err.Error(), "unknown theme") {
		t.Errorf("expected 'unknown theme' error, got: %v", err)
	}
}

func TestValidateTransition(t *testing.T) {
	for _, valid := range []string{"none", "fade", "slide", "convex", "concave", "zoom"} {
		if err := validateTransition(valid); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", valid, err)
		}
	}

	err := validateTransition("wipe")
	if err == nil {
		t.Fatal("expected error for unknown transition")
	}
	if !strings.Contains(err.Error(), "unknown transition") {
		t.Errorf("expected 'unknown transition' error, got: %v", err)
	}
}


