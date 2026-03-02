// Package assets embeds reveal.js and other static assets for the output site.
package assets

import "embed"

// RevealFS contains the reveal.js distribution files.
//
//go:embed reveal
var RevealFS embed.FS
