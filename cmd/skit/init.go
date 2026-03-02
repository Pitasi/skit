package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [path]",
		Short: "Create a starter presentation project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			return runInit(dir)
		},
	}
}

func runInit(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	wrote := 0

	// skit.md
	deckPath := filepath.Join(dir, "skit.md")
	if created, err := writeIfNotExists(deckPath, []byte(starterDeck)); err != nil {
		return err
	} else if created {
		wrote++
	} else {
		fmt.Printf("  skipped %s (already exists)\n", deckPath)
	}

	// themes/default/
	themeDir := filepath.Join(dir, "themes", "default")
	if err := os.MkdirAll(themeDir, 0o755); err != nil {
		return err
	}

	for _, f := range []struct {
		path    string
		content []byte
	}{
		{filepath.Join(themeDir, "theme.css"), []byte(starterThemeCSS)},
		{filepath.Join(themeDir, "theme.json"), []byte(starterThemeJSON)},
	} {
		if created, err := writeIfNotExists(f.path, f.content); err != nil {
			return err
		} else if created {
			wrote++
		} else {
			fmt.Printf("  skipped %s (already exists)\n", f.path)
		}
	}

	// .gitignore
	gitignorePath := filepath.Join(dir, ".gitignore")
	if created, err := writeIfNotExists(gitignorePath, []byte("dist/\n")); err != nil {
		return err
	} else if created {
		wrote++
	} else {
		fmt.Printf("  skipped %s (already exists)\n", gitignorePath)
	}

	if wrote == 0 {
		fmt.Println("All files already exist, nothing written.")
		return nil
	}

	fmt.Printf("Initialized project in %s\n", dir)
	fmt.Println("  skit.md              - your presentation")
	fmt.Println("  themes/default/      - theme files")
	fmt.Println("  .gitignore           - ignores dist/")
	fmt.Println()
	fmt.Println("Run 'skit build' to generate the static site.")
	return nil
}

// writeIfNotExists writes content to path only if the file does not already
// exist. Returns true if the file was created.
func writeIfNotExists(path string, content []byte) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}
	return true, os.WriteFile(path, content, 0o644)
}

const starterDeck = `---
title: My Presentation
author: Your Name
date: 2025-01-01
theme: white
transition: slide
aspectRatio: "16:9"
---

# Welcome

This is your first slide. The heading above is visible on the slide.

This paragraph is a speaker note — only you can see it during the presentation.

	This paragraph starts with a tab, so it appears on the slide.

---

# Second Slide

Here are your speaker notes for this slide. The audience won't see this.

:::slide
This block is also visible on the slide, using the :::slide directive.
:::

// This is a private comment — stripped from both slides and notes.

---

# Code Example

	` + "```" + `go
	func main() {
		fmt.Println("Hello, skit!")
	}
	` + "```" + `

This slide demonstrates a tab-promoted code block.
`

const starterThemeCSS = `/* Default skit theme */
.reveal {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-size: 42px;
}

.reveal h1, .reveal h2, .reveal h3 {
  font-weight: 700;
  text-transform: none;
  letter-spacing: -0.02em;
}

.reveal h1 {
  font-size: 2em;
}

.reveal h2 {
  font-size: 1.5em;
}

.reveal section img {
  max-width: 100%;
  max-height: 60vh;
  object-fit: contain;
  border: none;
  box-shadow: none;
}

.reveal pre {
  font-size: 0.55em;
  width: 100%;
}

.reveal code {
  font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace;
}
`

const starterThemeJSON = `{
  "name": "default",
  "author": "skit",
  "version": "1.0.0",
  "css": ["theme.css"]
}
`
