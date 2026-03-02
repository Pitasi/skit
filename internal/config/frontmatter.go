// Package config handles parsing of YAML front matter from deck files.
package config

import (
	"fmt"
	"strings"

	"github.com/Pitasi/skit/internal/model"
	"gopkg.in/yaml.v3"
)

// ParseFrontMatter extracts YAML front matter from the beginning of content.
// Returns the parsed Meta and the remaining body after the front matter.
// If no front matter is present, returns zero-value Meta and the full content.
func ParseFrontMatter(content string) (model.Meta, string, error) {
	// Strip UTF-8 BOM if present.
	content = strings.TrimPrefix(content, "\xef\xbb\xbf")

	trimmed := strings.TrimLeft(content, " \t\r\n")
	if !strings.HasPrefix(trimmed, "---") {
		return model.Meta{}, content, nil
	}

	// Find the opening delimiter line.
	startIdx := strings.Index(content, "---")
	afterOpen := content[startIdx+3:]

	// The opening --- must be followed by a newline.
	if len(afterOpen) == 0 || (afterOpen[0] != '\n' && afterOpen[0] != '\r') {
		return model.Meta{}, content, nil
	}
	afterOpen = afterOpen[1:]
	if len(afterOpen) > 0 && afterOpen[0] == '\n' && content[startIdx+3] == '\r' {
		afterOpen = afterOpen[1:]
	}

	// Find closing delimiter: --- or ...
	closeIdx := findClosingDelimiter(afterOpen)
	if closeIdx < 0 {
		return model.Meta{}, "", fmt.Errorf("unclosed YAML front matter: missing closing --- or ...")
	}

	yamlContent := afterOpen[:closeIdx]
	rest := afterOpen[closeIdx:]

	// Skip the closing delimiter line.
	rest = skipLine(rest)

	var meta model.Meta
	if err := yaml.Unmarshal([]byte(yamlContent), &meta); err != nil {
		return model.Meta{}, "", fmt.Errorf("parsing front matter YAML: %w", err)
	}

	return meta, rest, nil
}

// findClosingDelimiter finds the position of a closing --- or ... line.
// Returns the byte offset of the start of the closing delimiter line, or -1.
func findClosingDelimiter(s string) int {
	offset := 0
	for offset < len(s) {
		lineEnd := strings.IndexByte(s[offset:], '\n')
		var line string
		if lineEnd < 0 {
			line = s[offset:]
		} else {
			line = s[offset : offset+lineEnd]
		}
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" || trimmed == "..." {
			return offset
		}
		if lineEnd < 0 {
			break
		}
		offset += lineEnd + 1
	}
	return -1
}

// skipLine advances past the current line (including the newline).
func skipLine(s string) string {
	idx := strings.IndexByte(s, '\n')
	if idx < 0 {
		return ""
	}
	return s[idx+1:]
}
