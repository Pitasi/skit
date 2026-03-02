// Package render converts parsed deck content into HTML output.
package render

import (
	"bytes"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var _md = goldmark.New(
	goldmark.WithExtensions(
		extension.Table,
		extension.Strikethrough,
		extension.TaskList,
		highlighting.NewHighlighting(
			highlighting.WithStyle("monokai"),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(), // allow raw HTML in markdown
	),
)

// MarkdownToHTML converts a markdown string to HTML.
func MarkdownToHTML(md string) (string, error) {
	var buf bytes.Buffer
	if err := _md.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
