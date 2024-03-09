package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// FileEntry represents a file entry in the API response
type FileEntry struct {
	Label        string `json:"label"`
	Type         string `json:"type"`
	URL          string `json:"url"`
	AbsolutePath string `json:"absolutePath"`
}

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

// handleFile serves the content of the requested file, using AbsolutePath
func (w *Watcher) handleFile(wr http.ResponseWriter, r *http.Request) {

	// Find the file entry by URL
	w.filesLock.Lock()
	var fileEntry *FileEntry
	for _, entry := range w.files {
		if entry.URL == "http://localhost:9191"+r.URL.Path {
			fileEntry = &entry
			break
		}
	}
	w.filesLock.Unlock()

	// If the file entry was found, serve the file using its AbsolutePath
	if fileEntry != nil {
		http.ServeFile(wr, r, fileEntry.AbsolutePath)
		return
	}

	// If no file is found, return a 404 Not Found error
	http.NotFound(wr, r)
}

// Start begins watching the directories and serving the HTTP API
func (w *Watcher) Start() {
	for label, path := range w.config.directories {
		w.walkAndWatch(label, path)
	}

	go w.watch()

	http.HandleFunc("/registry.json", w.handleRegistry)
	http.HandleFunc("/files/", w.handleFile)
	log.Fatal(http.ListenAndServe(":9191", nil))
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
		URL:          fmt.Sprintf("http://localhost:9191/files/%s", filepath.Join(label, relativePath)),
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

// handleRegistry serves the registry.json endpoint
func (w *Watcher) handleRegistry(wr http.ResponseWriter, r *http.Request) {
	// print a message to the console
	fmt.Println("handleRegistry")

	w.filesLock.Lock()
	defer w.filesLock.Unlock()

	files := make([]FileEntry, 0, len(w.files))
	for _, fileEntry := range w.files {
		files = append(files, fileEntry)
	}

	jsonData, err := json.Marshal(files)
	if err != nil {
		http.Error(wr, "Failed to generate JSON", http.StatusInternalServerError)
		return
	}

	wr.Header().Set("Content-Type", "application/json")
	wr.Write(jsonData)
}

func main() {
	directoryFlag := flag.String("directory", "", "Directories to watch")
	extensionsFlag := flag.String("extensions", "", "File extensions to watch")
	flag.Parse()

	directories := parseDirectories(*directoryFlag)
	extensions := strings.Split(*extensionsFlag, ",")

	watcher := NewWatcher(WatcherConfig{
		directories: directories,
		extensions:  extensions,
	})

	watcher.Start()
}

// parseDirectories parses the directory flag into a map
func parseDirectories(flagValue string) map[string]string {
	directories := make(map[string]string)
	for _, dir := range strings.Split(flagValue, " ") {
		parts := strings.SplitN(dir, ":", 2)
		if len(parts) == 2 {
			directories[parts[0]] = parts[1]
		} else {
			directories[filepath.Base(parts[0])] = parts[0]
		}
	}
	return directories
}
