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

// TemplateSlide wraps a Slide with template.HTML fields for safe rendering.
type TemplateSlide struct {
	Index     int
	Title     string
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
</style>
</head>
<body>
<div class="reveal">
<div class="slides">
{{range .Slides}}
<section data-slide-index="{{.Index}}"{{if .Title}} data-title="{{.Title}}"{{end}}>
{{.SlideHTML}}
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
		tSlides[i] = TemplateSlide{
			Index:     s.Index,
			Title:     s.Title,
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
