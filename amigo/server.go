// server.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Define your HTTP server and handlers here
// For example, handleRegistry and handleFile functions

// handleRegistry serves the registry.json endpoint
func (w *Watcher) handleRegistry(wr http.ResponseWriter, r *http.Request) {
	// print a message to the console
	fmt.Println("handleRegistry ee")

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
