package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Pitasi/skit/internal/model"
)

var _headingRe = regexp.MustCompile(`^#{1,2}\s+(.+)`)

// ValidLayouts lists the layout names accepted by the :::layout directive.
var ValidLayouts = []string{
	"center", "split", "split-right", "split-3",
	"grid", "top-bottom", "background",
	"caption-left", "caption-right",
}

var _validLayoutSet = func() map[string]bool {
	m := make(map[string]bool, len(ValidLayouts))
	for _, l := range ValidLayouts {
		m[l] = true
	}
	return m
}()

var _layoutRe = regexp.MustCompile(`^:::layout\s+(\S+)$`)

// SeparateContent takes a raw slide chunk and separates it into slide-visible
// markdown and notes markdown, applying the promotion rules:
//
//  1. Lines starting with // are stripped entirely (private comments).
//  2. :::layout <name> sets the slide layout (stripped from output).
//  3. :::slide ... ::: blocks are extracted as slide-visible content.
//  4. The first heading (# or ##) becomes the slide title and is slide-visible.
//  5. Lines starting with a tab character are slide-visible (tab stripped).
//  6. Everything else is speaker notes.
//
// After collecting slide-visible lines, they are split into cells on blank-line
// boundaries.
func SeparateContent(chunk string, slideIndex int) (model.Slide, error) {
	lines := strings.Split(strings.TrimRight(chunk, "\n"), "\n")

	var (
		slideLines  []string
		notesLines  []string
		title       string
		layout      string
		mediaRefs   []string
		inDirective bool
		titleFound  bool
	)

	// addSlideBreak inserts a blank line between distinct content blocks
	// so that cell splitting can find the boundaries.
	addSlideBreak := func() {
		if len(slideLines) > 0 {
			slideLines = append(slideLines, "")
		}
	}

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Rule 1: strip // comments.
		if strings.HasPrefix(trimmed, "//") {
			i++
			continue
		}

		// Rule 2: :::layout directive.
		if m := _layoutRe.FindStringSubmatch(trimmed); m != nil {
			name := m[1]
			if !_validLayoutSet[name] {
				return model.Slide{}, fmt.Errorf(
					"slide %d: unknown layout %q; valid layouts: %s",
					slideIndex, name, strings.Join(ValidLayouts, ", "),
				)
			}
			if layout != "" {
				return model.Slide{}, fmt.Errorf(
					"slide %d: duplicate :::layout directive (already set to %q)",
					slideIndex, layout,
				)
			}
			layout = name
			i++
			continue
		}

		// Rule 3: :::slide directive blocks.
		if trimmed == ":::slide" {
			inDirective = true
			addSlideBreak()
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

		// Rule 4: first heading becomes title + slide-visible.
		if !titleFound {
			if m := _headingRe.FindStringSubmatch(trimmed); m != nil {
				title = m[1]
				titleFound = true
				addSlideBreak()
				slideLines = append(slideLines, line)
				i++
				continue
			}
		}

		// Rule 5: tab-promoted content.
		if strings.HasPrefix(line, "\t") {
			addSlideBreak()
			// Collect contiguous tab-promoted block.
			for i < len(lines) && strings.HasPrefix(lines[i], "\t") {
				promoted := strings.TrimPrefix(lines[i], "\t")
				slideLines = append(slideLines, promoted)
				collectMediaRef(promoted, &mediaRefs)
				i++
			}
			continue
		}

		// Rule 6: everything else is notes.
		notesLines = append(notesLines, line)
		collectMediaRef(line, &mediaRefs)
		i++
	}

	slideMD := strings.Join(slideLines, "\n")
	cells := splitCells(slideLines)

	return model.Slide{
		Index:         slideIndex,
		Title:         title,
		Layout:        layout,
		SlideMarkdown: slideMD,
		NotesMarkdown: strings.Join(notesLines, "\n"),
		Cells:         cells,
		MediaRefs:     mediaRefs,
	}, nil
}

// splitCells splits slide-visible lines into cells on blank-line boundaries.
// Each contiguous group of non-empty lines becomes one cell.
func splitCells(lines []string) []model.Cell {
	var cells []model.Cell
	var current []string

	flush := func() {
		if len(current) == 0 {
			return
		}
		cells = append(cells, model.Cell{
			Markdown: strings.Join(current, "\n"),
		})
		current = nil
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flush()
		} else {
			current = append(current, line)
		}
	}
	flush()

	return cells
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
