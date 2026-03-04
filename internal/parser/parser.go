package parser

import (
	"fmt"
	"os"

	"github.com/Pitasi/skit/internal/config"
	"github.com/Pitasi/skit/internal/model"
)

// Options controls parsing behavior.
type Options struct {
	SplitHeadings bool // additionally split slides on # and ## headings
}

// ParseFile reads a deck file and returns a fully parsed Deck (without HTML rendering).
func ParseFile(path string, opts Options) (*model.Deck, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading deck file: %w", err)
	}
	return Parse(string(data), opts)
}

// Parse parses deck content and returns a Deck with slide/notes markdown populated.
func Parse(content string, opts Options) (*model.Deck, error) {
	meta, body, err := config.ParseFrontMatter(content)
	if err != nil {
		return nil, err
	}

	chunks := SplitSlides(body)
	if opts.SplitHeadings {
		chunks = SplitByHeadings(chunks)
	}

	slides := make([]model.Slide, 0, len(chunks))
	for i, chunk := range chunks {
		// Validate before processing so malformed input is rejected early.
		if hasUnclosedDirective(chunk) {
			return nil, fmt.Errorf("slide %d: unclosed :::slide block", i)
		}

		slide, err := SeparateContent(chunk, i)
		if err != nil {
			return nil, err
		}
		slides = append(slides, slide)
	}

	return &model.Deck{
		Meta:   meta,
		Slides: slides,
	}, nil
}

func hasUnclosedDirective(chunk string) bool {
	opens := 0
	for _, line := range splitLines(chunk) {
		trimmed := trimSpace(line)
		if trimmed == ":::slide" {
			opens++
		} else if trimmed == ":::" && opens > 0 {
			opens--
		}
	}
	return opens > 0
}

func splitLines(s string) []string {
	// Avoid importing strings again; simple split.
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r') {
		i++
	}
	j := len(s)
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}
