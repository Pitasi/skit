package render

import (
	"fmt"
	"strings"

	"github.com/Pitasi/skit/internal/model"
)

// RenderDeck converts all slide and notes markdown to HTML in-place.
// Each cell's markdown is rendered individually so the template can
// wrap them in separate divs for layout.
func RenderDeck(deck *model.Deck) error {
	for i := range deck.Slides {
		s := &deck.Slides[i]

		// Render cells individually.
		var cellHTMLs []string
		for j := range s.Cells {
			c := &s.Cells[j]
			html, err := MarkdownToHTML(c.Markdown)
			if err != nil {
				return fmt.Errorf("slide %d cell %d: rendering markdown: %w", s.Index, j, err)
			}
			c.HTML = html
			cellHTMLs = append(cellHTMLs, html)
		}
		s.SlideHTML = strings.Join(cellHTMLs, "\n")

		notesHTML, err := MarkdownToHTML(s.NotesMarkdown)
		if err != nil {
			return fmt.Errorf("slide %d: rendering notes markdown: %w", s.Index, err)
		}
		s.NotesHTML = notesHTML
	}
	return nil
}
