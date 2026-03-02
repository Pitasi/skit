package main

import (
	"fmt"

	"github.com/Pitasi/skit/internal/parser"
	"github.com/Pitasi/skit/internal/site"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	var (
		inputFile     string
		outputDir     string
		theme         string
		baseURL       string
		aspectRatio   string
		transition    string
		splitHeadings bool
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the presentation into a static site",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(buildOpts{
				inputFile:     inputFile,
				outputDir:     outputDir,
				theme:         theme,
				baseURL:       baseURL,
				aspectRatio:   aspectRatio,
				transition:    transition,
				splitHeadings: splitHeadings,
			})
		},
	}

	cmd.Flags().StringVar(&inputFile, "in", "deck.md", "input markdown file")
	cmd.Flags().StringVar(&outputDir, "out", "dist", "output directory")
	cmd.Flags().StringVar(&theme, "theme", "", "built-in theme name, path to a .css file, or theme directory")
	cmd.Flags().StringVar(&baseURL, "base-url", "/", "base URL for assets")
	cmd.Flags().StringVar(&aspectRatio, "aspect", "", "aspect ratio (auto, 16:9, 4:3, 9:16, 1:1)")
	cmd.Flags().StringVar(&transition, "transition", "", "slide transition (none, fade, slide, convex, concave, zoom)")
	cmd.Flags().BoolVar(&splitHeadings, "split-headings", false, "also split slides on # and ## headings")

	return cmd
}

type buildOpts struct {
	inputFile     string
	outputDir     string
	theme         string
	baseURL       string
	aspectRatio   string
	transition    string
	splitHeadings bool
}

func runBuild(opts buildOpts) error {
	deck, err := parser.ParseFile(opts.inputFile, parser.Options{
		SplitHeadings: opts.splitHeadings,
	})
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	if err := site.Build(deck, site.BuildOptions{
		InputFile:   opts.inputFile,
		OutputDir:   opts.outputDir,
		Theme:       opts.theme,
		BaseURL:     opts.baseURL,
		AspectRatio: opts.aspectRatio,
		Transition:  opts.transition,
	}); err != nil {
		return fmt.Errorf("building site: %w", err)
	}

	fmt.Printf("Built %d slides → %s/\n", len(deck.Slides), opts.outputDir)
	return nil
}
