// Package parser handles splitting deck content into slides and separating
// slide-visible content from speaker notes.
package parser

import (
	"strings"
)

// SplitSlides splits the body (after front matter removal) into raw slide
// chunks using --- as the delimiter. It correctly ignores --- inside fenced
// code blocks (``` or ~~~).
func SplitSlides(body string) []string {
	var slides []string
	var current strings.Builder
	inFence := false
	fenceChar := byte(0)

	lines := strings.Split(body, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track fenced code blocks.
		if !inFence {
			if isOpeningFence(trimmed) {
				inFence = true
				fenceChar = trimmed[0]
				current.WriteString(line)
				current.WriteByte('\n')
				continue
			}
		} else {
			if isClosingFence(trimmed, fenceChar) {
				inFence = false
				fenceChar = 0
			}
			current.WriteString(line)
			current.WriteByte('\n')
			continue
		}

		// Outside a fence: check for slide delimiter.
		if trimmed == "---" {
			slides = append(slides, current.String())
			current.Reset()
			continue
		}

		current.WriteString(line)
		current.WriteByte('\n')
	}

	// Append the last slide.
	last := current.String()
	if len(strings.TrimSpace(last)) > 0 {
		slides = append(slides, last)
	}

	// If no slides were produced, return the whole body as one slide.
	if len(slides) == 0 && len(strings.TrimSpace(body)) > 0 {
		slides = append(slides, body)
	}

	return slides
}

// SplitByHeadings further splits slide chunks when encountering # or ##
// at the start of a line. Each heading starts a new sub-slide.
func SplitByHeadings(chunks []string) []string {
	var result []string
	for _, chunk := range chunks {
		result = append(result, splitChunkByHeadings(chunk)...)
	}
	return result
}

func splitChunkByHeadings(chunk string) []string {
	var slides []string
	var current strings.Builder
	inFence := false
	fenceChar := byte(0)
	lines := strings.Split(chunk, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track fenced code blocks so headings inside them are ignored.
		if !inFence {
			if isOpeningFence(trimmed) {
				inFence = true
				fenceChar = trimmed[0]
			}
		} else {
			if isClosingFence(trimmed, fenceChar) {
				inFence = false
				fenceChar = 0
			}
			current.WriteString(line)
			current.WriteByte('\n')
			continue
		}

		if !inFence && isHeadingSplit(trimmed) && current.Len() > 0 {
			text := current.String()
			if len(strings.TrimSpace(text)) > 0 {
				slides = append(slides, text)
			}
			current.Reset()
		}
		current.WriteString(line)
		current.WriteByte('\n')
	}

	last := current.String()
	if len(strings.TrimSpace(last)) > 0 {
		slides = append(slides, last)
	}

	return slides
}

func isHeadingSplit(trimmed string) bool {
	return strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ")
}

func isOpeningFence(trimmed string) bool {
	if len(trimmed) < 3 {
		return false
	}
	ch := trimmed[0]
	if ch != '`' && ch != '~' {
		return false
	}
	count := 0
	for _, b := range []byte(trimmed) {
		if b == ch {
			count++
		} else {
			break
		}
	}
	return count >= 3
}

func isClosingFence(trimmed string, fenceChar byte) bool {
	if len(trimmed) < 3 {
		return false
	}
	if trimmed[0] != fenceChar {
		return false
	}
	// A closing fence is only the fence characters (and optional whitespace).
	stripped := strings.TrimRight(trimmed, " \t")
	for _, b := range []byte(stripped) {
		if b != fenceChar {
			return false
		}
	}
	return len(stripped) >= 3
}
