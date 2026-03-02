# Deep research and Go CLI implementation plan inspired by iA Presenter

## How iA Presenter works in practice

iA Presenter is built around a “story-first” authoring workflow: you draft your talk as text, and only selectively “promote” pieces of that text to become slide-visible content. The marketing copy explicitly frames this as separating what you want to say (script) from what you want to show (slides), with the key behavior that “regular text is only visible to you” while “headlines show on the slide.” 

Under the hood, the app relies heavily on Markdown for formatting (bold/italic, lists, quotes, tables, code, math, footnotes, etc.). The crucial iA Presenter twist is its default visibility rule for *paragraphs*: normal paragraphs are treated as spoken text (speaker notes) by default, and only become slide-visible when you indent them. The Editor documentation makes the same distinction in UI terms: “Speech” content won’t appear on slides, and “Text on Slide” is what the audience sees; a fast way to make text visible is to put a tab character in front of it. 

Slide boundaries can be created in a few ways. In the Editor docs, a new slide can be created by typing `---` or entering two line breaks (and the Settings mention an option where pressing Return three times creates a new slide). The Markdown guide also documents horizontal rules (`---`) as the mechanism to split slides. 

A big part of the product value proposition is that layout and responsiveness are “automatic.” Presenter analyzes slide contents and chooses layouts accordingly, and it emphasizes that layouts are responsive and adapt to different screens and aspect ratios. The Layouts support page gives the concrete inputs it considers (number of visual blocks, types of graphics, heading level, order of blocks). The Features and Settings docs also describe aspect ratio options (responsive or fixed ratios like 4:3, 16:9, 9:16, 1:1), and note that fixed ratios also become defaults for exported documents. 

For delivery, iA Presenter has a dedicated “Presentation Mode” (teleprompter) that shows notes and slides simultaneously and supports “Speaker Notes Mode” and “Thumbnails Mode.” Navigation is designed around the “script-first” paradigm: you can navigate by notes (scrolling through note elements) or by slides (jumping title-to-title), with default keyboard mappings and customizable navigation keys; remote controllers are supported insofar as they map to keyboard-like navigation. 

For going beyond presenting, exports and sharing are central. The Export docs list Web Sharing, PDF, PowerPoint/PPTX (beta), images, HTML, and Markdown. For PDF, there are multiple layouts including options that incorporate speaker notes (e.g., “1 slide per page, speaker notes” and other handout-like layouts). For HTML export, iA describes exporting a full package including the presentation, graphics, theme, and a JavaScript rendering engine, which you can upload to hosts like GitHub Pages (possibly renaming “presentation.htm” to “index.html”). For “Web Sharing,” iA positions it as publishing a browser-viewable link; the Sharing docs clarify that web sharing is only available for licenses purchased directly (not via the App Store), links are effectively bearer-token URLs (“anybody who has the link can view”), and links remain available for one year by default, with storage quotas (e.g., 1 GB for subscription duration). 

Finally, theming is unusually “developer-friendly” for a presentation app: Presenter supports custom themes using HTML/CSS plus JSON metadata (`template.json`, `presets.json`), custom fonts (woff2 with `@font-face`), and a defined HTML structure with specific layout CSS classes. 

## Perks and flaws of iA Presenter as a model to copy

The most distinctive perk is the default-off text rule. By making paragraphs “spoken text” unless explicitly promoted (via indentation/tab), Presenter lets you write a real script without immediately cluttering slides. This supports rehearsing and delivery because the teleprompter mode is designed to foreground notes while still previewing slides. A second major perk is reduced “design thrash”: the product promises auto-layout and responsive slides across screens, with an explicit claim that it “automatically adapts” slides to different devices and aspect ratios. 

A third perk is the post-talk workflow: iA treats export/handout generation as first-class, offering multiple PDF layouts including speaker-notes variants and HTML export that’s described as an interactive, responsive package. This matters if your goal is a single source of truth that can become “slides + handout + web page.”

A fourth perk, relevant to your “edit in any editor” direction, is that the ecosystem around iA’s Markdown includes “Content Blocks,” a text-friendly transclusion syntax meant to embed images, CSVs, code files, and text files directly into Markdown-like documents. Even if Presenter itself has partial caveats (e.g., Import docs note it doesn’t support adding images as Content Blocks yet during import), the spec shows the same philosophy: keep authoring in plain text and make rich output an interpretation step. 

Flaws and constraints (as a blueprint to rebuild) fall into a few categories:

First, platform and ecosystem constraints: iA Presenter’s OS availability is not universal (macOS support is explicit, with iOS/iPadOS “coming soon” and Windows “TBD”), and web sharing is gated by licensing channel. If your goal is “edit anywhere” and “render anywhere,” you likely want a filesystem-first, cross-platform CLI that doesn’t depend on a specific app store or vendor service.

Second, portability and Markdown-compatibility trade-offs: iA offers settings like “single return starts a new paragraph,” while also warning that traditional Markdown expects two returns; the Markdown guide explicitly notes compatibility issues with other Markdown consumers when deviating from standard paragraph rules. A CLI MVP should avoid “nonstandard Markdown” defaults if you want clean interop with editors like Obsidian and any Markdown previewer.

Third, export and sharing limitations: PPTX export is explicitly described as beta with feature limitations, and web sharing links are time-limited and quota-limited by design. For a rebuild, you can decide whether to skip PPTX entirely (reasonable for MVP) and whether “sharing” is just “generate a static site” rather than a hosted service.

Fourth, layout complexity: iA’s auto-layout looks at multiple content factors and supports many layouts; reproducing that exact behavior is not an MVP task unless you’re ready to accept significant complexity and UX tuning. A rebuild should be explicit about what “responsive” means (scaling, reflow, or true adaptive layouts) and which subset is targeted first.

## Product spec for a Go CLI MVP that preserves the two non-negotiables

This section defines what the agent should implement. The MVP is explicitly a CLI-first system where the canonical source is plain files you can edit in any editor.

**Non-negotiable behaviors**

**Default-off text semantics (Presenter-notes first).**  
All plain paragraphs and most Markdown prose are treated as speaker notes by default, not slide-visible content—matching iA Presenter’s “normal paragraphs as spoken text” rule. The CLI must offer at least one explicit mechanism to “promote” content to slide-visible.

**Web rendering that works on different screens, and is printable as PDF.**  
The CLI must output a static site that works on phones/laptops/projectors, and there must be a reproducible PDF export story. This can be implemented by generating a reveal.js-based deck because reveal.js provides speaker notes views and a dedicated PDF export mode via a print stylesheet (noting it’s confirmed in Chrome/Chromium). 

**Must-have supporting features for a usable MVP foundation**

A speaker view (teleprompter-style) is essential; iA Presenter makes Presentation Mode a primary interface and reveal.js similarly supports a speaker view with a notes window and timer. 

Deterministic export outputs: iA’s HTML export produces a full package including theme and the rendering engine. The Go CLI should do the same: output directory contains `index.html` plus versioned assets, and can be uploaded to static hosting. 

A minimal theming mechanism: iA supports extensive HTML/CSS theming with structured layout classes. For MVP, provide a theme folder (CSS + optional assets + minimal JSON metadata), without trying to match iA’s full theme schema. 

A practical slide-splitting rule: accept `---` as a slide separator (aligned with iA’s docs). Also include an optional “split by headings” mode to mimic iA’s import rule that creates new slides for `#`/`##` headings. 

**Authoring format (filesystem)**

Define a single primary input file (for MVP) such as `deck.md`. Optionally accept a directory layout later.

The CLI should support:

- YAML front matter at the top for metadata (`title`, `author`, `date`, `theme`, `aspectRatio`, etc.). Be careful: YAML front matter uses `---` delimiters, which must not be misinterpreted as slide breaks.
- Slide separators: a line containing exactly `---` (optionally surrounded by whitespace) is the canonical slide delimiter. 
- “Promote to slide” markers: implement **both** mechanisms below so authors can choose ergonomics without losing compatibility:
  - **Tab-prefix promotion** (iA-like): any Markdown block whose lines begin with a literal tab (`\t`) is considered slide-visible; the tab is stripped before Markdown rendering. This matches iA’s “add a tab in front of your text” guidance. 
  - **Fenced directive promotion**: a custom block `:::slide … :::` that is rendered on the slide, while everything outside is notes by default. This avoids having to indent every line of a long paragraph and helps in editors that handle tabs awkwardly.
- “Note-only comments”: lines that begin with `//` are removed from both slide and audience output (useful for private TODOs). iA documents `//` as a comment mechanism that “only you can see.” 
- Headings: treat the first `#` or `##` inside each slide as the slide title by default (visible). This matches iA’s headline-on-slide framing and its import logic that splits slides by headings. 
- Media inclusion: support standard Markdown images (`![alt](path)`), and optionally support the Content Blocks spec as an extension (path/URL line + optional caption). The spec is designed for embedding images, CSV-as-table, code-as-codeblock, etc. 

## Implementation plan for the Go CLI

This is a concrete build plan intended to be handed to an implementing agent. It is opinionated and “production-grade MVP”: correctness, testability, and clear interfaces come first; perfect feature parity does not.

### Repository and module layout

Use a standard Go multi-package layout:

- `cmd/deck/`: Cobra/urfave CLI entrypoint (choose one and stick to it).
- `internal/config/`: parse config + front matter.
- `internal/parser/`: slide splitting + note/slide extraction + content blocks.
- `internal/render/`: Markdown → HTML, templating, theme resolution.
- `internal/site/`: output directory builder (assets copy, hashing, integrity).
- `internal/server/`: local dev server + live reload (optional but strongly recommended).
- `internal/pdf/`: PDF generation (headless Chrome) and/or scripted reveal PDF export mode.
- `internal/testdata/`: golden fixtures.

Keep exported APIs small; document packages with short package comments and public-symbol doc comments.

### CLI commands and flags

Implement these commands (names can be adjusted, but the shape should remain):

**`deck init [path]`**  
Creates a starter project:
- `deck.md` with example slide separators, notes, slide-visible blocks, and one image.
- `themes/default/` with `theme.css` and `theme.json`.
- `.gitignore` with `dist/`.

**`deck build`**  
Compiles `deck.md` into a static site in `dist/`:
- `dist/index.html`
- `dist/assets/...` (JS/CSS/fonts)
- `dist/media/...` (copied local media referenced in Markdown)

Flags:
- `--in deck.md`
- `--out dist`
- `--theme themes/default` (or theme name)
- `--base-url /` (important for GitHub Pages subpaths)
- `--aspect auto|16:9|4:3|9:16|1:1` (maps to reveal size or CSS) 
- `--notes-mode hidden|speaker|handout`  
  - `hidden`: audience deck (default).
  - `speaker`: enable reveal speaker view scaffolding.
  - `handout`: generate a parallel `handout.html` (notes + thumbnails).

**`deck serve`**  
Runs a local server:
- Serves `dist/` (build-once or build-on-change).
- Watches source files (markdown + theme + media) and rebuilds.
- Live reload in the browser via a small websocket script injection.

Flags:
- `--addr 127.0.0.1:8080`
- `--open` (open browser)
- `--watch` (default true)
- `--drafts` (if you add draft slide support later)

**`deck pdf`**  
Generates `dist/deck.pdf` (and optionally `dist/handout.pdf`).
Two acceptable MVP approaches:

- **Preferred**: use headless Chrome via Go (chromedp or similar) to open `dist/index.html?print-pdf` and call print-to-PDF. reveal.js documents `?print-pdf` as the path to PDF mode and notes inclusion via `showNotes`. 
- **Fallback**: shell out to a known tool (like DeckTape) only if installed; reveal.js mentions DeckTape as an alternative.  
  If you do this fallback, detect the binary and provide a crisp actionable error if missing.

Flags:
- `--in dist/index.html` (default from build output)
- `--out dist/deck.pdf`
- `--notes overlay|separate-page|off` (maps to reveal `showNotes`) 

### Parsing and document model

Define a stable internal model. Avoid coupling render output to raw markdown too early.

**Core structs (example)**  
- `Deck{ Meta, Slides[] }`
- `Meta{ Title, Author, Date, Theme, AspectRatio, BaseURL, Extra map[string]any }`
- `Slide{ Index, Title, SlideMarkdown, NotesMarkdown, SlideHTML, NotesHTML, MediaRefs[] }`

**Stage A: Load and normalise input**
- Read file as UTF-8.
- Strip BOM if present.
- Parse YAML front matter if present at file top (standard `---` … `---` or `---` … `...`).
- Preserve original line endings for better error reporting, but normalise to `\n` internally.

**Stage B: Split into slides**
- Implement a slide splitter that:
  - Recognises `---` as a slide delimiter only when it appears on its own line **and** not inside fenced code blocks.
  - Does not treat YAML front matter boundaries as slide splits.
- Add optional mode `--split headings`:
  - Within each slide chunk, split again when encountering `# ` or `## ` at the beginning of a line, mimicking iA’s import behavior (“new slide for each `# Heading1` or `## Heading2`”). 

**Stage C: Separate slide-visible vs notes content**
Do this per-slide chunk, before Markdown parsing.

Rules in priority order:
- Remove full-line comments beginning with `//` (drop from both outputs). 
- Extract `:::slide … :::` blocks into slide markdown (strip the wrapping markers).
- Everything else is notes **unless**:
  - It’s the first heading in the slide (becomes slide title/visible).
  - It is a “promoted” block via tab prefix:
    - A line begins with `\t` → considered slide-visible; strip leading `\t`.
    - Require promotion to be consistent across block lines:
      - For paragraphs, treat contiguous lines until blank line.
      - For lists, treat contiguous list section.
      - For fenced code blocks, if the opening fence is promoted, treat the whole fenced block as slide-visible and strip one leading tab from each line if present.
- Anything not captured as slide-visible remains notes markdown.

This approach mirrors iA’s “Speech vs Text on Slide,” but gives you a stable, deterministic rule in plain files. 

**Stage D: Content Blocks support**
Implement as an extension parser operating on both slide markdown and notes markdown:

- Detect lines that match the Content Blocks spec shape: a URL or local path, optionally indented by up to three spaces, optionally followed by a title in parentheses or quotes. 
- Resolve local paths relative to the markdown file directory.
- Convert to an explicit AST node in your model, or directly rewrite into Markdown that your renderer can handle:
  - For images: rewrite to `![caption](resolved-path)` or generate HTML `<figure>`.
  - For CSV: read file and convert to Markdown table (or HTML table).
  - For code/text files: embed content into fenced code block; infer language by extension (the spec maps extensions to languages via a JSON mapping). 
- Record `MediaRefs` for asset copying.

Keep this feature optional behind a flag in MVP if time is tight, but it is a strong “text-first” capability consistent with iA’s philosophy. 

### Rendering pipeline

**Markdown engine**
- Use a CommonMark-compliant Go Markdown library with extension support (tables, footnotes, strikethrough, etc.).
- Do not implement nonstandard “single-return paragraph” behavior; keep compatibility with general Markdown and editors. (iA explicitly treats single-return as an optional setting because traditional Markdown expects two Returns.) 

**Math**
- Option A (simple): keep `$...$` and `$$...$$` as-is and run KaTeX auto-render in the browser; iA uses KaTeX for math rendering. 
- Option B (server-side): render math during build (more complex; not needed for MVP).

**Syntax highlighting**
- Use client-side highlight.js or a Go highlighter. MVP recommendation: client-side to keep output consistent and avoid heavy parsing.

**HTML templating**
Generate `index.html` from templates with:
- reveal.js container elements.
- One `<section>` per slide containing slide HTML.
- One `<aside class="notes">` per slide containing notes HTML (speaker notes). reveal.js supports notes as `<aside>` and provides speaker view via plugin. 
- Slide title included in slide HTML (and optionally in `data-title` attributes for navigation).

**Theme and asset strategy**
- Bundle reveal.js + plugins locally into `dist/assets/reveal/` so the export is fully self-contained, similar to iA’s HTML export that includes the theme and a rendering engine in the export package. 
- Theme resolution:
  - `theme.json` with minimal fields: name, author, version, css file(s), optional fonts.
  - `theme.css` applied after reveal base theme.
- Copy local media referenced in markdown into `dist/media/` and rewrite URLs accordingly.
- Support `--base-url` for link rewriting (required for GitHub Pages subpaths). iA explicitly calls out potential renaming/hosting adjustments. 

### PDF generation approach

**Why reveal.js is a pragmatic MVP base**
- reveal.js documents a specific PDF export mode using `?print-pdf` and the browser print dialog, and notes it’s confirmed to work in Chrome/Chromium. 
- It also documents how to include speaker notes in PDF via `showNotes` and how to print notes on separate pages (`'separate-page'`). 

**Implementing `deck pdf`**
- Build the site to `dist/`.
- Launch headless Chrome:
  - Load `file:///.../dist/index.html?print-pdf`.
  - Wait for network idle and for reveal initialization.
  - Print to PDF with backgrounds enabled; set landscape by default.
- Add `--notes` support by toggling reveal config:
  - For overlay notes or separate-page notes, pass config through a small injected script or a templated config object in HTML.

### Speaker mode and navigation

For MVP, rely on reveal.js’s ecosystem for speaker view rather than re-implementing teleprompter UX from scratch.

- Enable speaker notes plugin and ensure notes are in `<aside class="notes">`. 
- Provide a `deck serve --speaker` flag that opens:
  - Audience view: `/`
  - Speaker view: either by instructing to press `S` (reveal default) or by opening `/speaker` that triggers it. reveal.js documents `S` as the shortcut. 
- Optional enhancement: implement “navigate by notes vs by slides” analogous to iA (notes-scroll vs title-jump). iA explicitly supports both navigation modes.  
  For MVP, reveal’s default navigation is slide-based; note-scroll mode can be added later as a custom plugin.

### Error handling and UX requirements

This MVP should be strict in ways that prevent silent wrong output.

- Fail fast on:
  - Missing input file
  - Unparseable front matter
  - Unclosed `:::slide` blocks
  - Missing local media references (provide file path + slide index)
- Provide deterministic slide indices and stable IDs for deep links.
- Provide `deck doctor` (optional but valuable): validates references, theme files, and that PDF prerequisites exist (Chrome detected).

## Testing, documentation, and quality gates

A production-grade foundation is mostly about repeatability.

**Unit tests**
- Slide splitting:
  - `---` separators
  - ignore `---` inside code fences
  - ignore YAML front matter separator
- Promotion extraction:
  - tab-promoted paragraphs/lists/code fences
  - `:::slide` blocks
  - first-heading-as-title behavior
  - `//` comment stripping 
- Content blocks:
  - local path parsing
  - caption parsing
  - CSV embedding and error cases
- Path rewriting and `--base-url` edge cases.

**Golden tests (snapshot-style)**
- For a set of `testdata/*.md`, assert that the generated `index.html` is byte-stable (after normalising build timestamps).
- For CSS/JS assets, assert file presence and expected directory structure.

**Integration tests**
- `deck build` end-to-end: run in temp dir, confirm outputs exist, confirm media copying.
- `deck serve`: start server on random port, fetch `GET /` returns 200 and contains slide content.
- `deck pdf` (optional in CI): run only when Chrome is available; otherwise skip with clear reason.

**Documentation**
- `README.md` should include:
  - Concept: default-off text and how to promote content
  - Minimal authoring example
  - How to export HTML and PDF
  - How to deploy to GitHub Pages (base URL and folder naming), paralleling iA’s “upload HTML export and possibly rename entry file” guidance. 

**CI**
- Run `go test ./...`
- Run `golangci-lint` (or similar)
- Build binaries for macOS/Linux/Windows
- Optional PDF test job gated on Chrome availability.

## Trade-offs, risks, and a realistic MVP scope boundary

This plan intentionally does **not** attempt to re-implement iA Presenter’s full “design engine” or its dozens of layouts; iA’s layout logic considers multiple factors and has many structured layout classes, and reproducing that faithfully would dominate the project. The MVP instead borrows a proven browser presentation runtime (reveal.js) and focuses engineering effort on what you actually asked for: default-off notes semantics and high-quality web/PDF export. 

Key risks to address up front:

- **Ambiguity of “explicitly added to slide.”** Tabs are explicit but can be awkward in some editors; the dual mechanism (tab + `:::slide`) is meant to keep authoring ergonomic without sacrificing plain-text portability. 
- **`---` collisions.** YAML front matter and code fences can contain `---`; you must treat slide splitting as a proper lexer problem, not “strings.Split.” 
- **PDF correctness.** reveal.js PDF export has documented expectations (Chrome/Chromium, `?print-pdf`, print settings) that must be matched in headless mode. 
- **Scope creep into hosted sharing.** iA’s sharing includes quotas, link lifetimes, and licensing gates.  
  For MVP, “sharing” should mean “generate static output that you host wherever you want,” not “build a service.”
