package render

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/Pitasi/skit/internal/model"
)

// ValidTransitions lists the reveal.js transition styles accepted by the tool.
var ValidTransitions = []string{
	"none", "fade", "slide", "convex", "concave", "zoom",
}

var _validTransitionSet = func() map[string]bool {
	m := make(map[string]bool, len(ValidTransitions))
	for _, t := range ValidTransitions {
		m[t] = true
	}
	return m
}()

// TemplateCell wraps a Cell with template.HTML for safe rendering.
type TemplateCell struct {
	HTML template.HTML
}

// TemplateSlide wraps a Slide with template.HTML fields for safe rendering.
type TemplateSlide struct {
	Index     int
	Title     string
	Layout    string
	Cells     []TemplateCell
	SlideHTML template.HTML
	NotesHTML template.HTML
}

// TemplateData is the data passed to the HTML template.
type TemplateData struct {
	Meta        model.Meta
	Slides      []TemplateSlide
	BaseURL     string
	AspectRatio string
	Transition  string // "none", "fade", "slide", "convex", "concave", "zoom"
}

const _indexTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Meta.Title}}</title>
{{if .Meta.Author}}<meta name="author" content="{{.Meta.Author}}">{{end}}
<link rel="stylesheet" href="{{.BaseURL}}assets/reveal/dist/reset.css">
<link rel="stylesheet" href="{{.BaseURL}}assets/reveal/dist/reveal.css">
<link rel="stylesheet" href="{{.BaseURL}}assets/reveal/plugin/highlight/monokai.css">
<link rel="stylesheet" href="{{.BaseURL}}assets/theme.css" id="theme">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/katex.min.css">
<style>
.reveal .slides section { text-align: left; }

/* Layout: cells container */
.reveal .slides section .cells {
  display: flex;
  flex-direction: column;
  height: 100%;
}

/* center: all cells centered */
.reveal .slides section.layout-center .cells {
  justify-content: center;
  align-items: center;
  text-align: center;
}

/* split: two columns 50/50 */
.reveal .slides section.layout-split .cells {
  flex-direction: row;
  align-items: center;
  gap: 2rem;
}
.reveal .slides section.layout-split .cells > .cell {
  flex: 1;
  min-width: 0;
}

/* split-right: two columns, reversed order */
.reveal .slides section.layout-split-right .cells {
  flex-direction: row-reverse;
  align-items: center;
  gap: 2rem;
}
.reveal .slides section.layout-split-right .cells > .cell {
  flex: 1;
  min-width: 0;
}

/* split-3: three equal columns */
.reveal .slides section.layout-split-3 .cells {
  flex-direction: row;
  align-items: center;
  gap: 2rem;
}
.reveal .slides section.layout-split-3 .cells > .cell {
  flex: 1;
  min-width: 0;
}

/* grid: 2-column auto-flow grid */
.reveal .slides section.layout-grid .cells {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  align-items: center;
}

/* top-bottom: two rows */
.reveal .slides section.layout-top-bottom .cells {
  justify-content: center;
  gap: 2rem;
}
.reveal .slides section.layout-top-bottom .cells > .cell {
  flex: 1;
  min-height: 0;
}

/* background: first cell is full-bleed behind the rest */
.reveal .slides section.layout-background .cells {
  position: relative;
  height: 100%;
}
.reveal .slides section.layout-background .cells > .cell:first-child {
  position: absolute;
  inset: 0;
  z-index: 0;
}
.reveal .slides section.layout-background .cells > .cell:first-child img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.reveal .slides section.layout-background .cells > .cell:not(:first-child) {
  position: relative;
  z-index: 1;
}

/* caption-left: image ~65% right, text ~35% left */
.reveal .slides section.layout-caption-left .cells {
  flex-direction: row;
  align-items: center;
  gap: 2rem;
}
.reveal .slides section.layout-caption-left .cells > .cell:first-child {
  flex: 0 0 35%;
}
.reveal .slides section.layout-caption-left .cells > .cell:nth-child(2) {
  flex: 1;
  min-width: 0;
}

/* caption-right: image ~65% left, text ~35% right */
.reveal .slides section.layout-caption-right .cells {
  flex-direction: row;
  align-items: center;
  gap: 2rem;
}
.reveal .slides section.layout-caption-right .cells > .cell:first-child {
  flex: 1;
  min-width: 0;
}
.reveal .slides section.layout-caption-right .cells > .cell:nth-child(2) {
  flex: 0 0 35%;
}

/* Ensure images in cells fill their container */
.reveal .slides section .cell img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
</style>
</head>
<body>
<div class="reveal">
<div class="slides">
{{range .Slides}}
<section data-slide-index="{{.Index}}"{{if .Layout}} class="layout-{{.Layout}}"{{end}}{{if .Title}} data-title="{{.Title}}"{{end}}>
{{if .Cells}}<div class="cells">
{{range .Cells}}<div class="cell">
{{.HTML}}
</div>
{{end}}</div>
{{end}}
{{if .NotesHTML}}<aside class="notes">
{{.NotesHTML}}
</aside>{{end}}
</section>
{{end}}
</div>
</div>
<script src="{{.BaseURL}}assets/reveal/dist/reveal.js"></script>
<script src="{{.BaseURL}}assets/reveal/plugin/notes/notes.js"></script>
<script src="{{.BaseURL}}assets/reveal/plugin/highlight/highlight.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/katex.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/contrib/auto-render.min.js"
  onload="renderMathInElement(document.body, {delimiters: [{left: '$$', right: '$$', display: true},{left: '$', right: '$', display: false}]});">
</script>
<script>
Reveal.initialize({
  hash: true,
  {{if eq .AspectRatio "16:9"}}width: 1920, height: 1080,{{end}}
  {{if eq .AspectRatio "4:3"}}width: 1024, height: 768,{{end}}
  {{if eq .AspectRatio "9:16"}}width: 1080, height: 1920,{{end}}
  {{if eq .AspectRatio "1:1"}}width: 1080, height: 1080,{{end}}
  {{if .Transition}}transition: '{{.Transition}}',{{end}}
  plugins: [ RevealNotes, RevealHighlight ]
});
</script>
</body>
</html>`

var _tmpl = template.Must(template.New("index").Parse(_indexTemplate))

// NewTemplateData creates TemplateData from model types, converting HTML strings
// to template.HTML for safe rendering.
func NewTemplateData(meta model.Meta, slides []model.Slide, baseURL, aspectRatio, transition string) TemplateData {
	tSlides := make([]TemplateSlide, len(slides))
	for i, s := range slides {
		tCells := make([]TemplateCell, len(s.Cells))
		for j, c := range s.Cells {
			tCells[j] = TemplateCell{HTML: template.HTML(c.HTML)}
		}
		tSlides[i] = TemplateSlide{
			Index:     s.Index,
			Title:     s.Title,
			Layout:    s.Layout,
			Cells:     tCells,
			SlideHTML: template.HTML(s.SlideHTML),
			NotesHTML: template.HTML(s.NotesHTML),
		}
	}
	return TemplateData{
		Meta:        meta,
		Slides:      tSlides,
		BaseURL:     baseURL,
		AspectRatio: aspectRatio,
		Transition:  transition,
	}
}

// RenderHTML generates the final index.html content from a rendered deck.
func RenderHTML(data TemplateData) (string, error) {
	if data.BaseURL == "" {
		data.BaseURL = "/"
	}
	// Ensure trailing slash.
	if data.BaseURL[len(data.BaseURL)-1] != '/' {
		data.BaseURL += "/"
	}
	if data.AspectRatio == "" {
		data.AspectRatio = "auto"
	}
	// Guard against injection: only allow known transition values in the
	// JS output. This protects callers who use RenderHTML directly without
	// going through site.Build's validation.
	if data.Transition != "" && !_validTransitionSet[data.Transition] {
		return "", fmt.Errorf("unknown transition %q", data.Transition)
	}

	var buf bytes.Buffer
	if err := _tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
