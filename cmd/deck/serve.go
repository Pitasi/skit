package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var (
		addr          string
		inputFile     string
		outputDir     string
		themeDir      string
		baseURL       string
		aspectRatio   string
		notesMode     string
		splitHeadings bool
		watch         bool
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run a local dev server with live reload",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(serveOpts{
				addr:          addr,
				inputFile:     inputFile,
				outputDir:     outputDir,
				themeDir:      themeDir,
				baseURL:       baseURL,
				aspectRatio:   aspectRatio,
				notesMode:     notesMode,
				splitHeadings: splitHeadings,
				watch:         watch,
			})
		},
	}

	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8080", "listen address")
	cmd.Flags().StringVar(&inputFile, "in", "deck.md", "input markdown file")
	cmd.Flags().StringVar(&outputDir, "out", "dist", "output directory")
	cmd.Flags().StringVar(&themeDir, "theme", "", "theme directory")
	cmd.Flags().StringVar(&baseURL, "base-url", "/", "base URL for assets")
	cmd.Flags().StringVar(&aspectRatio, "aspect", "", "aspect ratio")
	cmd.Flags().StringVar(&notesMode, "notes-mode", "speaker", "notes mode (hidden, speaker, handout)")
	cmd.Flags().BoolVar(&splitHeadings, "split-headings", false, "split slides on headings")
	cmd.Flags().BoolVar(&watch, "watch", true, "watch for file changes")

	return cmd
}

type serveOpts struct {
	addr          string
	inputFile     string
	outputDir     string
	themeDir      string
	baseURL       string
	aspectRatio   string
	notesMode     string
	splitHeadings bool
	watch         bool
}

func runServe(opts serveOpts) error {
	// Initial build.
	if err := runBuild(buildOpts{
		inputFile:     opts.inputFile,
		outputDir:     opts.outputDir,
		themeDir:      opts.themeDir,
		baseURL:       opts.baseURL,
		aspectRatio:   opts.aspectRatio,
		notesMode:     opts.notesMode,
		splitHeadings: opts.splitHeadings,
	}); err != nil {
		return err
	}

	// Live reload hub.
	hub := newReloadHub()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var wg sync.WaitGroup

	// File watcher.
	if opts.watch {
		wg.Add(1)
		go func() {
			defer wg.Done()
			watchAndRebuild(ctx, opts, hub)
		}()
	}

	// Resolve output directory to an absolute path once, fail early on error.
	absOutDir, err := filepath.Abs(opts.outputDir)
	if err != nil {
		return fmt.Errorf("resolving output dir: %w", err)
	}

	// HTTP server.
	mux := http.NewServeMux()

	// Serve static files with live-reload script injection.
	fileServer := http.FileServer(http.Dir(opts.outputDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Inject live-reload script into HTML responses.
		if r.URL.Path == "/" || strings.HasSuffix(r.URL.Path, ".html") {
			// Resolve and validate the path stays within outputDir.
			name := r.URL.Path
			if name == "/" {
				name = "/index.html"
			}
			cleaned := filepath.Clean(filepath.Join(absOutDir, name))
			if !strings.HasPrefix(cleaned, absOutDir+string(filepath.Separator)) && cleaned != absOutDir {
				http.NotFound(w, r)
				return
			}
			data, err := os.ReadFile(cleaned)
			if err != nil {
				fileServer.ServeHTTP(w, r)
				return
			}
			html := strings.Replace(string(data), "</body>", liveReloadScript+"</body>", 1)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, html)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	// WebSocket endpoint for live reload.
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.handleWS(w, r)
	})

	listener, err := net.Listen("tcp", opts.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	fmt.Printf("Serving at http://%s\n", listener.Addr())
	fmt.Println("Press Ctrl+C to stop.")

	server := &http.Server{Handler: mux}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}

	wg.Wait()
	return nil
}

const liveReloadScript = `<script>
(function() {
  var ws = new WebSocket("ws://" + location.host + "/ws");
  ws.onmessage = function(e) {
    if (e.data === "reload") location.reload();
  };
  ws.onclose = function() {
    setTimeout(function() { location.reload(); }, 1000);
  };
})();
</script>
`

func watchAndRebuild(ctx context.Context, opts serveOpts, hub *reloadHub) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("watch error: %v", err)
		return
	}
	defer watcher.Close()

	// Watch the input file and theme directory.
	watcher.Add(opts.inputFile)
	if opts.themeDir != "" {
		watcher.Add(opts.themeDir)
	}
	// Also watch the directory containing the input file for new media.
	watcher.Add(filepath.Dir(opts.inputFile))

	// Debounce rebuilds.
	var debounce *time.Timer
	for {
		select {
		case <-ctx.Done():
			if debounce != nil {
				debounce.Stop()
			}
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.AfterFunc(200*time.Millisecond, func() {
					fmt.Println("Rebuilding...")
					if err := runBuild(buildOpts{
						inputFile:     opts.inputFile,
						outputDir:     opts.outputDir,
						themeDir:      opts.themeDir,
						baseURL:       opts.baseURL,
						aspectRatio:   opts.aspectRatio,
						notesMode:     opts.notesMode,
						splitHeadings: opts.splitHeadings,
					}); err != nil {
						log.Printf("rebuild error: %v", err)
						return
					}
					hub.broadcast("reload")
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watch error: %v", err)
		}
	}
}

// reloadHub manages WebSocket connections for live reload.
type reloadHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

var _upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func newReloadHub() *reloadHub {
	return &reloadHub{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (h *reloadHub) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := _upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()

	// Keep connection alive; remove on close.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()
			return
		}
	}
}

func (h *reloadHub) broadcast(msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}
