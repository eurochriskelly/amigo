// watcher.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// WatcherConfig holds the watcher configuration
type WatcherConfig struct {
	directories map[string]string
	extensions  []string
}

// Watcher holds the data needed for the file watcher
type Watcher struct {
	config    WatcherConfig
	watcher   *fsnotify.Watcher
	files     map[string]FileEntry
	filesLock sync.Mutex
}

// NewWatcher creates a new watcher instance
// Include the methods related to the Watcher here (e.g., Start, addFile, etc.)

// NewWatcher creates a new watcher instance
func NewWatcher(config WatcherConfig) *Watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	return &Watcher{
		config:  config,
		watcher: w,
		files:   make(map[string]FileEntry),
	}
}

// Start begins watching the directories and serving the HTTP API
func (w *Watcher) Start() {
	for label, path := range w.config.directories {
		w.walkAndWatch(label, path)
	}

	go w.watch()
	http.HandleFunc("/registry.json", w.handleRegistry)
	http.HandleFunc("/files/", w.handleFile)
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}

// walkAndWatch walks through directories and sets up watchers
func (w *Watcher) walkAndWatch(label, path string) {
	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && w.isWatchedExtension(filePath) {
			relativePath, _ := filepath.Rel(path, filePath)
			w.addFile(label, relativePath, filePath)
		}
		return nil
	})

	w.watcher.Add(path)
}

// isWatchedExtension checks if the file has one of the watched extensions
func (w *Watcher) isWatchedExtension(filePath string) bool {
	for _, ext := range w.config.extensions {
		if strings.HasSuffix(filePath, "."+ext) {
			return true
		}
	}
	return false
}

// addFile adds a file to the internal map
func (w *Watcher) addFile(label, relativePath, filePath string) {
	ext := filepath.Ext(filePath)
	w.filesLock.Lock()
	defer w.filesLock.Unlock()
	w.files[filePath] = FileEntry{
		Label:        filepath.Join(label, strings.TrimSuffix(relativePath, ext)),
		Type:         ext[1:], // remove the dot
		URL:          fmt.Sprintf("http://localhost:%s/files/%s", PORT, filepath.Join(label, relativePath)),
		AbsolutePath: filePath,
	}
}

// watch handles file system notifications
func (w *Watcher) watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("modified file:", event.Name)
				// Update the files map if necessary
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}
