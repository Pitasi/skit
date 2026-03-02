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

// Slide represents a single slide with its visible and notes content.
type Slide struct {
	Index         int
	Title         string
	SlideMarkdown string
	NotesMarkdown string
	SlideHTML     string
	NotesHTML     string
	MediaRefs     []string
}
