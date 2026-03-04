// Package site assembles the output directory with HTML, assets, and media.
package site

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Pitasi/skit/internal/assets"
	"github.com/Pitasi/skit/internal/model"
	"github.com/Pitasi/skit/internal/render"
)

// BuildOptions configures the site build.
type BuildOptions struct {
	InputFile   string
	OutputDir   string
	Theme       string
	BaseURL     string
	AspectRatio string
	Transition  string
}

// Build generates the static site from a parsed (but not yet rendered) deck.
// It rewrites media paths in the markdown, renders to HTML, then assembles
// the output directory.
func Build(deck *model.Deck, opts BuildOptions) error {
	outDir := opts.OutputDir
	if outDir == "" {
		outDir = "dist"
	}

	// Clean and create output directory.
	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("cleaning output dir: %w", err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	// Copy reveal.js assets.
	if err := copyRevealAssets(outDir); err != nil {
		return fmt.Errorf("copying reveal assets: %w", err)
	}

	// Resolve theme: CLI flag takes priority, then front-matter, then default.
	theme := opts.Theme
	if theme == "" {
		theme = deck.Meta.Theme
	}
	if err := resolveTheme(outDir, theme); err != nil {
		return fmt.Errorf("copying theme: %w", err)
	}

	// Copy media files referenced in slides.
	inputDir := filepath.Dir(opts.InputFile)
	if err := copyMedia(deck, inputDir, outDir); err != nil {
		return fmt.Errorf("copying media: %w", err)
	}

	// Rewrite media paths in markdown before rendering to HTML.
	// This ensures only actual image references are rewritten, not
	// arbitrary text that happens to match a filename.
	rewriteMediaPathsInMarkdown(deck, opts.BaseURL)

	// Render markdown to HTML.
	if err := render.RenderDeck(deck); err != nil {
		return fmt.Errorf("rendering markdown: %w", err)
	}

	// Render final HTML page.
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = "/"
	}
	aspectRatio := opts.AspectRatio
	if aspectRatio == "" {
		aspectRatio = deck.Meta.AspectRatio
	}

	transition := opts.Transition
	if transition == "" {
		transition = deck.Meta.Transition
	}
	if transition != "" {
		if err := validateTransition(transition); err != nil {
			return err
		}
	}

	data := render.NewTemplateData(deck.Meta, deck.Slides, baseURL, aspectRatio, transition)
	html, err := render.RenderHTML(data)
	if err != nil {
		return fmt.Errorf("rendering HTML: %w", err)
	}

	indexPath := filepath.Join(outDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(html), 0o644); err != nil {
		return fmt.Errorf("writing index.html: %w", err)
	}

	return nil
}

func copyRevealAssets(outDir string) error {
	return fs.WalkDir(assets.RevealFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dest := filepath.Join(outDir, "assets", path)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		data, err := assets.RevealFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
}

// BuiltinThemes lists the reveal.js theme names embedded in the binary.
var BuiltinThemes = []string{
	"beige", "black", "blood", "dracula", "league",
	"moon", "night", "serif", "simple", "sky",
	"solarized", "white",
}

// resolveTheme picks the theme CSS and writes it to <outDir>/assets/theme.css.
//
// Resolution order:
//  1. Empty → default ("white" built-in).
//  2. Built-in name (e.g. "moon") → embedded CSS.
//  3. Path to a .css file → copy it.
//  4. Directory containing theme.css → copy theme.css from it.
//  5. Otherwise → error.
func resolveTheme(outDir, theme string) error {
	themeDest := filepath.Join(outDir, "assets", "theme.css")

	if theme == "" {
		theme = "white"
	}

	// Try built-in name.
	builtinPath := "reveal/dist/theme/" + theme + ".css"
	if data, err := assets.RevealFS.ReadFile(builtinPath); err == nil {
		return os.WriteFile(themeDest, data, 0o644)
	}

	// Try as a .css file path.
	if strings.HasSuffix(theme, ".css") {
		if _, err := os.Stat(theme); err == nil {
			return copyFile(theme, themeDest)
		}
		return fmt.Errorf("theme CSS file not found: %s", theme)
	}

	// Try as a directory containing theme.css.
	dirCSS := filepath.Join(theme, "theme.css")
	if _, err := os.Stat(dirCSS); err == nil {
		return copyFile(dirCSS, themeDest)
	}

	return fmt.Errorf("unknown theme %q; available built-in themes: %s",
		theme, strings.Join(BuiltinThemes, ", "))
}

func validateTransition(t string) error {
	for _, v := range render.ValidTransitions {
		if v == t {
			return nil
		}
	}
	return fmt.Errorf("unknown transition %q; valid values: %s",
		t, strings.Join(render.ValidTransitions, ", "))
}



func copyMedia(deck *model.Deck, inputDir, outDir string) error {
	mediaDir := filepath.Join(outDir, "media")
	seen := make(map[string]bool)

	absInputDir, err := filepath.Abs(inputDir)
	if err != nil {
		return fmt.Errorf("resolving input dir: %w", err)
	}

	for _, slide := range deck.Slides {
		for _, ref := range slide.MediaRefs {
			if seen[ref] {
				continue
			}
			seen[ref] = true

			src := filepath.Clean(filepath.Join(inputDir, ref))
			absSrc, err := filepath.Abs(src)
			if err != nil {
				return fmt.Errorf("slide %d: resolving media path %q: %w", slide.Index, ref, err)
			}
			// Reject paths that escape the input directory.
			if !strings.HasPrefix(absSrc, absInputDir+string(filepath.Separator)) && absSrc != absInputDir {
				return fmt.Errorf("slide %d: media ref %q resolves outside input directory", slide.Index, ref)
			}

			if _, err := os.Stat(src); err != nil {
				return fmt.Errorf("slide %d: missing media file %q", slide.Index, ref)
			}

			dest := filepath.Join(mediaDir, ref)
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			if err := copyFile(src, dest); err != nil {
				return err
			}
		}
	}
	return nil
}

// rewriteMediaPathsInMarkdown rewrites local media references in the markdown
// source (before HTML rendering). Only paths inside Markdown image syntax
// ![...](path) are rewritten, avoiding false matches in prose text.
func rewriteMediaPathsInMarkdown(deck *model.Deck, baseURL string) {
	if baseURL == "" {
		baseURL = "/"
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	for i := range deck.Slides {
		s := &deck.Slides[i]
		refs := make(map[string]bool, len(s.MediaRefs))
		for _, ref := range s.MediaRefs {
			refs[ref] = true
		}
		rewrite := func(md string) string {
			return _imgRefRe.ReplaceAllStringFunc(md, func(match string) string {
				sub := _imgRefRe.FindStringSubmatch(match)
				if len(sub) < 3 {
					return match
				}
				path := sub[2]
				if !refs[path] {
					return match
				}
				newPath := baseURL + "media/" + path
				return strings.Replace(match, "]("+path+")", "]("+newPath+")", 1)
			})
		}
		s.SlideMarkdown = rewrite(s.SlideMarkdown)
		for j := range s.Cells {
			s.Cells[j].Markdown = rewrite(s.Cells[j].Markdown)
		}
		s.NotesMarkdown = rewrite(s.NotesMarkdown)
	}
}

// _imgRefRe matches Markdown image syntax: ![alt](path)
var _imgRefRe = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

func copyFile(src, dst string) (retErr error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()

	_, err = io.Copy(out, in)
	return err
}
