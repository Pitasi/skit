// Package model defines the document model for a presentation deck.
package model

// Deck is the top-level representation of a parsed presentation.
type Deck struct {
	Meta   Meta
	Slides []Slide
}

// Meta holds front-matter metadata for the deck.
type Meta struct {
	Title       string         `yaml:"title"`
	Author      string         `yaml:"author"`
	Date        string         `yaml:"date"`
	Theme       string         `yaml:"theme"`
	Transition  string         `yaml:"transition"`
	AspectRatio string         `yaml:"aspectRatio"`
	BaseURL     string         `yaml:"baseUrl"`
	Extra       map[string]any `yaml:",inline"`
}

// Cell is a block of slide-visible content. Cells are separated by blank
// lines in the slide's visible markdown and are the unit of layout.
type Cell struct {
	Markdown string
	HTML     string
}

// Slide represents a single slide with its visible and notes content.
type Slide struct {
	Index         int
	Title         string
	Layout        string // layout directive, e.g. "split", "background"
	SlideMarkdown string
	NotesMarkdown string
	SlideHTML     string
	NotesHTML     string
	Cells         []Cell
	MediaRefs     []string
}
