// Package serve builds a site once, serves the output over HTTP, and
// rebuilds automatically as source files change.
package serve

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/handbitesdog/marge/internal/build"
)

// debounce is how long Run waits after the last filesystem event before
// rebuilding, since editors typically fire several events per save.
const debounce = 150 * time.Millisecond

// watchDirs are the SrcDir subdirectories watched for changes; anything
// outside them (e.g. DistDir itself) is ignored.
var watchDirs = []string{"components", "pages", "content", "static"}

// Run builds the site at src into dist, serves dist over HTTP on addr, and
// rebuilds whenever a file under src's components/, pages/, content/, or
// static/ subtrees changes. It blocks until interrupted (SIGINT), then
// shuts down cleanly.
func Run(src, dist, addr string) error {
	opts := build.Options{SrcDir: src, DistDir: dist}
	if err := build.Run(opts); err != nil {
		return fmt.Errorf("initial build: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	for _, sub := range watchDirs {
		if err := watchTree(watcher, filepath.Join(src, sub)); err != nil {
			return err
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %q: %w", addr, err)
	}

	server := &http.Server{Handler: http.FileServer(http.Dir(dist))}
	serverErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("serve %q: %w", addr, err)
		}
	}()

	log.Printf("serving %s at http://localhost:%d", dist, listener.Addr().(*net.TCPAddr).Port)

	go watchLoop(ctx, watcher, opts)

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

// watchTree adds a watch for root and every directory beneath it (fsnotify
// only watches one level per Add). It is a no-op if root does not exist, so
// an absent static/ directory behaves the same way it does for CopyStatic.
func watchTree(watcher *fsnotify.Watcher, root string) error {
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		return watcher.Add(p)
	})
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("watch %q: %w", root, err)
	}
	return nil
}

// watchLoop rebuilds opts on filesystem events, debounced so a burst of
// events from a single save triggers one rebuild. A newly created directory
// is watched immediately so future edits inside it are picked up too. A
// rebuild error is logged, not returned, so a bad edit can't kill the
// server.
func watchLoop(ctx context.Context, watcher *fsnotify.Watcher, opts build.Options) {
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	rebuild := func() {
		if err := build.Run(opts); err != nil {
			log.Printf("rebuild failed: %v", err)
		} else {
			log.Println("rebuilt")
		}
	}

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if err := watchTree(watcher, event.Name); err != nil {
						log.Printf("watch %q: %v", event.Name, err)
					}
				}
			}
			if timer == nil {
				timer = time.AfterFunc(debounce, rebuild)
			} else {
				timer.Reset(debounce)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watch error: %v", err)
		}
	}
}
