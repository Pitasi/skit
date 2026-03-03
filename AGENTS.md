# Project Guidelines

## Setup

```bash
go mod download      # fetch dependencies
go run ./cmd/skit    # run without building a binary
```

Go 1.25+ required (see `go.mod`).

## Commands

- `go mod download -json MODULE` — returns the `Dir` path for inspecting dependency source.
- `go doc foo.Bar` / `go doc -all foo` — read documentation for packages, types, functions.
- `go run ./cmd/skit` — build and run the CLI without leaving binary artifacts.

CLI subcommands: `init`, `build`, `serve`, `pdf` (see `cmd/skit/`).

## Testing

```bash
go test ./...                                  # run all tests
go test ./internal/render/                     # run render tests only
go test ./internal/parser/                     # run parser tests only
UPDATE_GOLDEN=1 go test ./internal/render/     # regenerate golden file
```

- Golden file comparison: `internal/render/render_test.go` vs `internal/testdata/basic.golden.html`.
- Integration test: `internal/integration_test.go` — builds a full presentation end-to-end.
- Test fixtures live in `internal/testdata/`.

## Lint / Format

```bash
gofmt -w .           # format all Go files
go vet ./...         # static analysis
```

Run both before committing.

## Commit Conventions

Based on repo history: imperative mood, concise subject line describing the change.

Examples from the log:
- `Rename CLI binary from 'deck' to 'skit'`
- `Add theme selection and slide transition options`
- `Remove unused NotesMode field and --notes-mode flag`

No CI workflows configured. No branch naming convention enforced.

## Code Style

Follow `STYLE.md` (Uber Go Style Guide). Key points:
- Use field names in struct literals; omit zero-value fields.
- Group imports: stdlib, then external.
- Prefer `strconv` over `fmt` for conversions.
- Reduce nesting; use early returns.
- Table-driven tests.

## Architecture

`skit` converts Markdown files into reveal.js slide decks.

### Pipeline

1. **Parse** (`internal/parser/`) — reads `.md`, splits on `---`, extracts front-matter (`internal/config/`), produces `model.Deck`.
2. **Render** (`internal/render/`) — converts slide markdown to HTML via goldmark, assembles final `index.html` from Go template.
3. **Build** (`internal/site/`) — copies reveal.js assets, theme CSS, media files, rewrites media paths, calls render, writes `dist/`.

### Key Directories

```
cmd/skit/              CLI entry point (cobra commands)
internal/
  assets/              Embedded reveal.js files (embed.FS)
  config/              Front-matter parsing (YAML)
  model/               Deck, Meta, Slide types
  parser/              Markdown splitting and slide/notes separation
  render/              Goldmark markdown-to-HTML, HTML template
  site/                Build orchestration, asset copying, media path rewriting
  testdata/            Test fixtures (basic.md, basic.golden.html)
```

### Key Packages

- `model` — `Deck`, `Meta`, `Slide`. Slides carry markdown, rendered HTML, and `MediaRefs`.
- `parser` — `splitter.go` splits on `---`; `separator.go` separates slide vs. notes (tab-promoted lines and `:::slide` directives go on-slide, rest becomes speaker notes).
- `render` — `markdown.go` configures goldmark (table, strikethrough, task-list extensions). Syntax highlighting is client-side via reveal.js highlight.js plugin. `template.go` produces the final HTML page.
- `assets` — embeds `reveal/dist/` and `reveal/plugin/` via `embed.FS`.
- `site` — `builder.go` copies assets, resolves media paths (with directory-escape protection), rewrites markdown image refs, calls render.
