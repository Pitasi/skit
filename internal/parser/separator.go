package parser

import (
	"regexp"
	"strings"

	"github.com/Pitasi/skit/internal/model"
)

var _headingRe = regexp.MustCompile(`^#{1,2}\s+(.+)`)

// SeparateContent takes a raw slide chunk and separates it into slide-visible
// markdown and notes markdown, applying the promotion rules:
//
//  1. Lines starting with // are stripped entirely (private comments).
//  2. :::slide ... ::: blocks are extracted as slide-visible content.
//  3. The first heading (# or ##) becomes the slide title and is slide-visible.
//  4. Lines starting with a tab character are slide-visible (tab stripped).
//  5. Everything else is speaker notes.
func SeparateContent(chunk string, slideIndex int) model.Slide {
	lines := strings.Split(strings.TrimRight(chunk, "\n"), "\n")

	var (
		slideLines []string
		notesLines []string
		title      string
		mediaRefs  []string
		inDirective bool
		titleFound  bool
	)

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Rule 1: strip // comments.
		if strings.HasPrefix(trimmed, "//") {
			i++
			continue
		}

		// Rule 2: :::slide directive blocks.
		if trimmed == ":::slide" {
			inDirective = true
			i++
			continue
		}
		if inDirective {
			if trimmed == ":::" {
				inDirective = false
				i++
				continue
			}
			slideLines = append(slideLines, line)
			collectMediaRef(line, &mediaRefs)
			i++
			continue
		}

		// Rule 3: first heading becomes title + slide-visible.
		if !titleFound {
			if m := _headingRe.FindStringSubmatch(trimmed); m != nil {
				title = m[1]
				titleFound = true
				slideLines = append(slideLines, line)
				i++
				continue
			}
		}

		// Rule 4: tab-promoted content.
		if strings.HasPrefix(line, "\t") {
			// Collect contiguous tab-promoted block.
			for i < len(lines) && strings.HasPrefix(lines[i], "\t") {
				promoted := strings.TrimPrefix(lines[i], "\t")
				slideLines = append(slideLines, promoted)
				collectMediaRef(promoted, &mediaRefs)
				i++
			}
			continue
		}

		// Handle tab-promoted fenced code blocks: if a tab-prefixed line
		// opens a fence, consume until the fence closes.
		// (Already handled above since we consume contiguous tab lines.)

		// Rule 5: everything else is notes.
		notesLines = append(notesLines, line)
		collectMediaRef(line, &mediaRefs)
		i++
	}

	return model.Slide{
		Index:         slideIndex,
		Title:         title,
		SlideMarkdown: strings.Join(slideLines, "\n"),
		NotesMarkdown: strings.Join(notesLines, "\n"),
		MediaRefs:     mediaRefs,
	}
}

var _imgRe = regexp.MustCompile(`!\[.*?\]\((.+?)\)`)

func collectMediaRef(line string, refs *[]string) {
	matches := _imgRe.FindAllStringSubmatch(line, -1)
	for _, m := range matches {
		path := m[1]
		// Skip URLs.
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			continue
		}
		*refs = append(*refs, path)
	}
}
