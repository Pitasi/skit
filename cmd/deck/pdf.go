package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
)

func newPDFCmd() *cobra.Command {
	var (
		distDir   string
		outputPDF string
		notes     string
	)

	cmd := &cobra.Command{
		Use:   "pdf",
		Short: "Generate PDF from the built presentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPDF(pdfOpts{
				distDir:   distDir,
				outputPDF: outputPDF,
				notes:     notes,
			})
		},
	}

	cmd.Flags().StringVar(&distDir, "dist", "dist", "build output directory to serve")
	cmd.Flags().StringVar(&outputPDF, "out", "dist/deck.pdf", "output PDF file")
	cmd.Flags().StringVar(&notes, "notes", "off", "notes mode (overlay, separate-page, off)")

	return cmd
}

type pdfOpts struct {
	distDir   string
	outputPDF string
	notes     string
}

func runPDF(opts pdfOpts) error {
	indexPath := filepath.Join(opts.distDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return fmt.Errorf("input file not found: %s (run 'deck build' first)", indexPath)
	}

	// Serve the dist directory over HTTP so that asset paths (e.g.
	// /assets/reveal/dist/reveal.css) resolve correctly. The file://
	// protocol can't resolve root-relative paths against the dist folder.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting local server: %w", err)
	}
	defer listener.Close()

	server := &http.Server{Handler: http.FileServer(http.Dir(opts.distDir))}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	defer server.Close()

	// Fail fast if the server couldn't start.
	select {
	case err := <-serveErr:
		return fmt.Errorf("local server failed: %w", err)
	default:
	}

	url := fmt.Sprintf("http://%s/?print-pdf", listener.Addr())

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
