package main

import (
	"fmt"

	"github.com/Pitasi/skit/internal/parser"
	"github.com/Pitasi/skit/internal/site"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	var (
		inputFile   string
		outputDir   string
		themeDir    string
		baseURL     string
		aspectRatio string
		notesMode   string
		splitHeadings bool
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the presentation into a static site",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(buildOpts{
				inputFile:     inputFile,
				outputDir:     outputDir,
				themeDir:      themeDir,
				baseURL:       baseURL,
				aspectRatio:   aspectRatio,
				notesMode:     notesMode,
				splitHeadings: splitHeadings,
			})
		},
	}

	cmd.Flags().StringVar(&inputFile, "in", "deck.md", "input markdown file")
	cmd.Flags().StringVar(&outputDir, "out", "dist", "output directory")
	cmd.Flags().StringVar(&themeDir, "theme", "", "theme directory")
	cmd.Flags().StringVar(&baseURL, "base-url", "/", "base URL for assets")
	cmd.Flags().StringVar(&aspectRatio, "aspect", "", "aspect ratio (auto, 16:9, 4:3, 9:16, 1:1)")
	cmd.Flags().StringVar(&notesMode, "notes-mode", "hidden", "notes mode (hidden, speaker, handout)")
	cmd.Flags().BoolVar(&splitHeadings, "split-headings", false, "also split slides on # and ## headings")

	return cmd
}

type buildOpts struct {
	inputFile     string
	outputDir     string
	themeDir      string
	baseURL       string
	aspectRatio   string
	notesMode     string
	splitHeadings bool
}

func runBuild(opts buildOpts) error {
	deck, err := parser.ParseFile(opts.inputFile, parser.Options{
		SplitHeadings: opts.splitHeadings,
	})
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	// Resolve theme directory from meta if not specified via flag.
	themeDir := opts.themeDir
	if themeDir == "" && deck.Meta.Theme != "" {
		themeDir = "themes/" + deck.Meta.Theme
	}

	if err := site.Build(deck, site.BuildOptions{
		InputFile:   opts.inputFile,
		OutputDir:   opts.outputDir,
		ThemeDir:    themeDir,
		BaseURL:     opts.baseURL,
		AspectRatio: opts.aspectRatio,
		NotesMode:   opts.notesMode,
	}); err != nil {
		return fmt.Errorf("building site: %w", err)
	}

	fmt.Printf("Built %d slides → %s/\n", len(deck.Slides), opts.outputDir)
	return nil
}
