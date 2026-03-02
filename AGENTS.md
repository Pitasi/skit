# Project Guidelines

## Commands
- `go mod download -json MODULE` - Returns the Dir path that can be used to see source file of a dependency, and answer questions about a dependency. 
- `go doc foo.Bar` and `go doc -all foo` - Read documentation for packages, types, functions, etc.
- `go run .` or `go run ./cmd/foo` - Build a program and run it, without leaving binary artifacts behind.

## Testing
- `go test ./...` — run all tests.
- `UPDATE_GOLDEN=1 go test ./internal/render/` — regenerate the golden file at `internal/testdata/basic.golden.html`.
- Golden file comparison lives in `internal/render/render_test.go`.
- Integration test in `internal/integration_test.go` builds a full presentation end-to-end.

## Code Style

Refer to `STYLE.md` for the Go source code you write.

## Architecture

`skit` converts Markdown files into reveal.js slide decks.

### CLI (`cmd/skit/`)
Entry point. Subcommands: `init`, `build`, `serve`, `pdf`. Uses cobra.

### Pipeline
1. **Parse** (`internal/parser/`) — reads a `.md` file, splits on `---` separators, extracts front-matter (`internal/config/`), and produces a `model.Deck`.
2. **Render** (`internal/render/`) — converts each slide's markdown to HTML via goldmark, then assembles the final `index.html` from a Go template.
3. **Build** (`internal/site/`) — orchestrates the full output: copies reveal.js assets, theme CSS, media files, rewrites media paths, calls render, writes `dist/`.

### Key packages
- `internal/model/` — `Deck`, `Meta`, `Slide` types. Slides carry both markdown and rendered HTML fields, plus `MediaRefs`.
- `internal/parser/` — `splitter.go` splits on `---`, `separator.go` separates slide vs. notes content (tab-promoted lines and `:::slide` directives go on-slide, the rest becomes speaker notes).
- `internal/render/markdown.go` — goldmark instance with table, strikethrough, task-list extensions. Syntax highlighting is handled client-side by reveal.js's highlight.js plugin (not server-side).
- `internal/render/template.go` — the HTML template that produces the final page. Loads reveal.js, highlight.js plugin, KaTeX, and a theme CSS.
- `internal/assets/` — embeds reveal.js distribution files (`reveal/dist/`, `reveal/plugin/`) via `embed.FS`.
- `internal/site/builder.go` — copies assets, resolves media paths (with directory-escape protection), rewrites markdown image refs, then calls render.

### Test data
- `internal/testdata/basic.md` — fixture used by render and parser tests.
- `internal/testdata/basic.golden.html` — expected full HTML output for golden-file comparison.