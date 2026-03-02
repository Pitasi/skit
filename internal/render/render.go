package render

import (
	"fmt"

	"github.com/Pitasi/skit/internal/model"
)

// RenderDeck converts all slide and notes markdown to HTML in-place.
func RenderDeck(deck *model.Deck) error {
	for i := range deck.Slides {
		s := &deck.Slides[i]

		slideHTML, err := MarkdownToHTML(s.SlideMarkdown)
		if err != nil {
			return fmt.Errorf("slide %d: rendering slide markdown: %w", s.Index, err)
		}
		s.SlideHTML = slideHTML

		notesHTML, err := MarkdownToHTML(s.NotesMarkdown)
		if err != nil {
			return fmt.Errorf("slide %d: rendering notes markdown: %w", s.Index, err)
		}
		s.NotesHTML = notesHTML
	}
	return nil
}
