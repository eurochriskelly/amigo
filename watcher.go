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
	Label string `json:"label"`
	Type  string `json:"type"`
	URL   string `json:"url"`
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

// handleFile serves the content of the requested file
func (w *Watcher) handleFile(wr http.ResponseWriter, r *http.Request) {
	// Extract the label and filename from the URL path
	urlPath := strings.TrimPrefix(r.URL.Path, "/files/")
	pathParts := strings.SplitN(urlPath, "/", 3) // Split into ["label", "filename.extension"]

	// print out variables
	fmt.Printf("urlPath: %s\n", urlPath)
	fmt.Printf("pathParts: %s\n", pathParts)

	if len(pathParts) < 2 {
		http.Error(wr, "Invalid file request", http.StatusBadRequest)
		return
	}

	label := pathParts[0]
	fileName := pathParts[1]

	// Reconstruct the relative or absolute file path
	filePath, err := w.getFilePathFromLabelAndName(label, fileName)
	if err != nil {
		http.NotFound(wr, r)
		return
	}

	// Serve the file
	http.ServeFile(wr, r, filePath)
}

// getFilePathFromLabelAndName finds the filesystem path for a given label and filename
func (w *Watcher) getFilePathFromLabelAndName(label, fileName string) (string, error) {
	w.filesLock.Lock()
	defer w.filesLock.Unlock()

	for _, entry := range w.files {
		fmt.Sprintf("%s", entry)
		if strings.HasPrefix(entry.URL, fmt.Sprintf("http://localhost:9191/files/%s/%s", label, fileName)) {
			return entry.Label, nil // Note: This assumes `entry.Label` stores the correct file system path
		}
	}

	return "", fmt.Errorf("file not found")
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
		Label: filepath.Join(label, strings.TrimSuffix(relativePath, ext)),
		Type:  ext[1:], // remove the dot
		URL:   fmt.Sprintf("http://localhost:9191/files/%s", filepath.Join(label, relativePath)),
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
