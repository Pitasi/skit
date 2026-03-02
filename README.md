# skit

A CLI for building presentations from Markdown. Inspired by iA Presenter's "default-off" text model: paragraphs are speaker notes unless explicitly promoted to slides.

## Install

```
go install github.com/Pitasi/skit/cmd/deck@latest
```

## Quick start

```bash
deck init my-talk
cd my-talk
deck build
deck serve
```

## Authoring format

Write your presentation in a single Markdown file (`deck.md`). YAML front matter at the top sets metadata:

```markdown
---
title: My Talk
author: Jane Doe
date: 2025-06-01
theme: default
aspectRatio: "16:9"
---
```

### Slide separators

Use `---` on its own line to separate slides. `---` inside fenced code blocks is ignored.

### Default-off text

All paragraphs are **speaker notes** by default. To make content visible on the slide, use one of:

**Tab promotion** — prefix lines with a tab character:

```markdown
# Slide Title

This is a speaker note.

	This paragraph appears on the slide (tab-prefixed).
```

**Directive block** — wrap content in `:::slide ... :::`:

```markdown
:::slide
This block appears on the slide.
:::

This is a speaker note.
```

**Headings** — the first `#` or `##` in each slide is automatically visible as the slide title.

### Comments

Lines starting with `//` are stripped from both slides and notes:

```markdown
// TODO: add a better example here
```

### Images

Standard Markdown images work:

```markdown
	![Photo](images/photo.png)
```

Local images are copied to the output directory automatically.

## Commands

### `deck init [path]`

Creates a starter project with `deck.md`, a default theme, and `.gitignore`.

### `deck build`

Compiles `deck.md` into a static site in `dist/`.

| Flag | Default | Description |
|------|---------|-------------|
| `--in` | `deck.md` | Input file |
| `--out` | `dist` | Output directory |
| `--theme` | (from front matter) | Theme directory |
| `--base-url` | `/` | Base URL for assets |
| `--aspect` | (from front matter) | Aspect ratio: `16:9`, `4:3`, `9:16`, `1:1` |
| `--notes-mode` | `hidden` | `hidden`, `speaker`, or `handout` |
| `--split-headings` | `false` | Also split slides on `#`/`##` headings |

### `deck serve`

Runs a local dev server with live reload. Watches for file changes and rebuilds automatically.

| Flag | Default | Description |
|------|---------|-------------|
| `--addr` | `127.0.0.1:8080` | Listen address |
| `--watch` | `true` | Watch for changes |

Press `S` in the browser to open the speaker notes view.

### `deck pdf`

Generates a PDF using headless Chrome. Requires Chrome/Chromium installed.

```bash
deck build
deck pdf
```

| Flag | Default | Description |
|------|---------|-------------|
| `--in` | `dist/index.html` | Input HTML |
| `--out` | `dist/deck.pdf` | Output PDF |
| `--notes` | `off` | `overlay`, `separate-page`, or `off` |

## Theming

Themes live in `themes/<name>/` with:

- `theme.css` — CSS applied after the reveal.js base theme
- `theme.json` — metadata (name, author, version, css files, fonts)

## Deploying to GitHub Pages

```bash
deck build --base-url /my-repo/
```

Upload the `dist/` directory contents to your GitHub Pages branch.

## Running tests

```bash
go test ./...
```
