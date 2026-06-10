# skit

A CLI for building presentations from Markdown. Inspired by iA Presenter's "default-off" text model: paragraphs are speaker notes unless explicitly promoted to slides.

## Install

```
go install github.com/Pitasi/skit/cmd/skit@latest
```

## Quick start

```bash
skit init my-talk
cd my-talk
skit build
skit serve
```

## Authoring format

Write your presentation in a single Markdown file (`skit.md`). YAML front matter at the top sets metadata:

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

### `skit init [path]`

Creates a starter project with `skit.md`, a default theme, and `.gitignore`.

### `skit build`

Compiles `skit.md` into a static site in `dist/`.

| Flag | Default | Description |
|------|---------|-------------|
| `--in` | `skit.md` | Input file |
| `--out` | `dist` | Output directory |
| `--theme` | (from front matter) | Built-in name, `.css` file path, or theme directory |
| `--base-url` | `/` | Base URL for assets |
| `--aspect` | (from front matter) | Aspect ratio: `16:9`, `4:3`, `9:16`, `1:1` |
| `--transition` | (from front matter) | Slide transition: `none`, `fade`, `slide`, `convex`, `concave`, `zoom` |
| `--split-headings` | `false` | Also split slides on `#`/`##` headings |

### `skit serve`

Runs a local dev server with live reload. Watches for file changes and rebuilds automatically.

| Flag | Default | Description |
|------|---------|-------------|
| `--addr` | `127.0.0.1:8080` | Listen address |
| `--in` | `skit.md` | Input file |
| `--out` | `dist` | Output directory |
| `--theme` | (from front matter) | Built-in name, `.css` file path, or theme directory |
| `--base-url` | `/` | Base URL for assets |
| `--aspect` | (from front matter) | Aspect ratio |
| `--transition` | (from front matter) | Slide transition |
| `--split-headings` | `false` | Split slides on headings |
| `--watch` | `true` | Watch for changes |

Press `S` in the browser to open the speaker notes view.

### `skit pdf`

Generates a PDF using headless Chrome. Requires Chrome/Chromium installed.

```bash
skit build
skit pdf
```

| Flag | Default | Description |
|------|---------|-------------|
| `--in` | `dist/index.html` | Input HTML |
| `--out` | `dist/skit.pdf` | Output PDF |
| `--notes` | `off` | `overlay`, `separate-page`, or `off` |

## Theming

skit supports three ways to set a theme, in order of priority:

1. **CLI flag** — `--theme <value>` on `build` or `serve`
2. **Front matter** — `theme:` field in the YAML header
3. **Default** — `white` (a built-in reveal.js theme)

The `--theme` value is resolved as follows:

| Value | Example | What happens |
|-------|---------|--------------|
| Built-in name | `--theme dracula` | Uses the bundled reveal.js theme CSS |
| Path to `.css` file | `--theme ./my-theme.css` | Copies that file as the theme |
| Directory | `--theme ./themes/default` | Copies `theme.css` from that directory |

### Built-in themes

These are the standard reveal.js themes bundled with skit:

`beige` · `black` · `blood` · `dracula` · `league` · `moon` · `night` · `serif` · `simple` · `sky` · `solarized` · `white`

Use them by name in front matter or with `--theme`:

```yaml
---
theme: dracula
---
```

### Custom themes

To create a custom theme, write a CSS file that styles the `.reveal` container. The CSS is loaded _after_ the reveal.js base styles, so you can override anything.

A minimal custom theme:

```css
.reveal {
  font-family: "Georgia", serif;
  font-size: 40px;
  color: #333;
}

.reveal h1, .reveal h2, .reveal h3 {
  color: #1a1a2e;
  text-transform: none;
}

.reveal pre {
  font-size: 0.55em;
}
```

Use it directly:

```bash
skit build --theme ./my-theme.css
skit serve --theme ./my-theme.css
```

Or put it in a directory and reference the directory:

```
themes/
  my-theme/
    theme.css
    theme.json   # optional metadata
```

```bash
skit build --theme ./themes/my-theme
```

You can also set it in front matter by path:

```yaml
---
theme: ./themes/my-theme
---
```

### Scaffolded theme

`skit init` creates a starter theme at `themes/default/` with `theme.css` and `theme.json`. Edit `themes/default/theme.css` to customize fonts, colors, and layout. The `theme.json` file is metadata only — it is not read by the build pipeline.

### Live reload

When using `skit serve --theme <path>`, changes to the theme file or directory are watched and trigger an automatic rebuild.

## Deploying to GitHub Pages

```bash
skit build --base-url /my-repo/
```

Upload the `dist/` directory contents to your GitHub Pages branch.

## Running tests

```bash
go test ./...
```
