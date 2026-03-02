package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
)

func newPDFCmd() *cobra.Command {
	var (
		inputHTML string
		outputPDF string
		notes     string
	)

	cmd := &cobra.Command{
		Use:   "pdf",
		Short: "Generate PDF from the built presentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPDF(pdfOpts{
				inputHTML: inputHTML,
				outputPDF: outputPDF,
				notes:     notes,
			})
		},
	}

	cmd.Flags().StringVar(&inputHTML, "in", "dist/index.html", "input HTML file")
	cmd.Flags().StringVar(&outputPDF, "out", "dist/deck.pdf", "output PDF file")
	cmd.Flags().StringVar(&notes, "notes", "off", "notes mode (overlay, separate-page, off)")

	return cmd
}

type pdfOpts struct {
	inputHTML string
	outputPDF string
	notes     string
}

func runPDF(opts pdfOpts) error {
	absPath, err := filepath.Abs(opts.inputHTML)
	if err != nil {
		return fmt.Errorf("resolving input path: %w", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("input file not found: %s (run 'deck build' first)", opts.inputHTML)
	}

	url := "file://" + absPath + "?print-pdf"

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set a timeout for the whole operation.
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var pdfBuf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		// Wait for reveal.js to initialize.
		chromedp.WaitReady(".reveal .slides"),
		chromedp.Sleep(2*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithLandscape(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBuf = buf
			return nil
		}),
	); err != nil {
		return fmt.Errorf("generating PDF: %w\n\nMake sure Chrome/Chromium is installed.", err)
	}

	if err := os.MkdirAll(filepath.Dir(opts.outputPDF), 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(opts.outputPDF, pdfBuf, 0o644); err != nil {
		return err
	}

	fmt.Printf("PDF generated: %s\n", opts.outputPDF)
	return nil
}
