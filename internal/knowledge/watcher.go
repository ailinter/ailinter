package knowledge

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher monitors source directories for file changes and incrementally
// updates the knowledge graph.
type FileWatcher struct {
	graph    *Graph
	watcher  *fsnotify.Watcher
	repoRoot string
	watched  map[string]bool // paths currently being watched
	done     chan struct{}
}

// NewFileWatcher creates a new file watcher for the given graph and repository root.
func NewFileWatcher(g *Graph, repoRoot string) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create fsnotify watcher: %w", err)
	}

	fw := &FileWatcher{
		graph:    g,
		watcher:  w,
		repoRoot: repoRoot,
		watched:  make(map[string]bool),
		done:     make(chan struct{}),
	}

	// Watch the internal source directory and other relevant paths
	watchedPaths := []string{
		filepath.Join(repoRoot, "internal"),
	}

	for _, p := range watchedPaths {
		if err := fw.addRecursive(p); err != nil {
			Logf("warning: could not watch %s: %v", p, err)
		}
	}

	return fw, nil
}

func (fw *FileWatcher) addRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		if !info.IsDir() {
			return nil
		}
		// Skip hidden directories, vendor, node_modules
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") || base == "vendor" || base == "node_modules" || base == "testdata" {
			return filepath.SkipDir
		}
		if strings.Contains(path, "/.git") || strings.Contains(path, "\\.git") {
			return filepath.SkipDir
		}
		return fw.addDir(path)
	})
}

func (fw *FileWatcher) addDir(path string) error {
	if fw.watched[path] {
		return nil
	}
	if err := fw.watcher.Add(path); err != nil {
		return err
	}
	fw.watched[path] = true
	return nil
}

// Start begins watching for file changes. Blocks until context is cancelled.
func (fw *FileWatcher) Start(ctx context.Context) error {
	Logf("file watcher started for %s", fw.repoRoot)

	defer fw.watcher.Close()

	for {
		select {
		case <-ctx.Done():
			Logf("file watcher shutting down")
			return nil

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return nil
			}

			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return nil
			}
			Logf("file watcher error: %v", err)
		}
	}
}

// Stop gracefully stops the file watcher.
func (fw *FileWatcher) Stop() error {
	close(fw.done)
	return fw.watcher.Close()
}

func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	ext := filepath.Ext(event.Name)

	switch {
	case event.Op&fsnotify.Create != 0:
		fw.handleCreate(event.Name, ext)
	case event.Op&fsnotify.Write != 0:
		fw.handleWrite(event.Name, ext)
	case event.Op&fsnotify.Remove != 0:
		fw.handleRemove(event.Name, ext)
	case event.Op&fsnotify.Rename != 0:
		fw.handleRemove(event.Name, ext)
	}
}

func (fw *FileWatcher) handleCreate(path, ext string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		fw.addRecursive(path)
		return
	}

	// Debounce: wait a tiny bit for writes to settle
	time.Sleep(100 * time.Millisecond)
	fw.reprocessFile(path, ext)
}

func (fw *FileWatcher) handleWrite(path, ext string) {
	// Debounce rapid writes
	time.Sleep(200 * time.Millisecond)
	fw.reprocessFile(path, ext)
}

func (fw *FileWatcher) handleRemove(path, ext string) {
	// When a file is removed, remove its nodes and edges from the graph
	if ext != ".go" && ext != ".md" {
		return
	}

	relPath, err := filepath.Rel(fw.repoRoot, path)
	if err != nil {
		relPath = path
	}

	nodeID := slug("file", relPath)
	fw.graph.RemoveEdgesForNode(nodeID)
	fw.graph.RemoveNode(nodeID)
	Logf("removed nodes for deleted file: %s", relPath)
}

func (fw *FileWatcher) reprocessFile(path, ext string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}

	relPath, err := filepath.Rel(fw.repoRoot, path)
	if err != nil {
		relPath = path
	}

	fw.graph.RemoveEdgesForNode(slug("file", relPath))
	fw.graph.RemoveNode(slug("file", relPath))

	switch ext {
	case ".go":
		if _, err := ingestGoFile(fw.graph, fw.repoRoot, path); err != nil {
			Logf("warning: could not ingest %s: %v", relPath, err)
			return
		}
		Logf("re-ingested Go file: %s", relPath)

	case ".md":
		if err := ingestSpecFile(fw.graph, path); err != nil {
			Logf("warning: could not ingest spec %s: %v", relPath, err)
			return
		}
		Logf("re-ingested spec file: %s", relPath)

	default:
		return
	}

	// Persist snapshot after each incremental update
	fw.graph.mu.RLock()
	fw.graph.LastBuilt = time.Now()
	fw.graph.SourceFiles[path] = time.Now()
	fw.graph.mu.RUnlock()

	if err := fw.graph.ExportJSON(fw.graph.snapshotPath()); err != nil {
		Logf("warning: could not persist snapshot after file change: %v", err)
	}
}
