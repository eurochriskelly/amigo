// server.go
package main

import (
	"encoding/json"
	"net/http"
)

func enableCors(w *http.ResponseWriter) {
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
    (*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    (*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// handleRegistry serves the registry.json endpoint
func (w *Watcher) handleRegistry(wr http.ResponseWriter, r *http.Request) {
  enableCors(&wr)
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
  enableCors(&wr)
	// Find the file entry by URL
	w.filesLock.Lock()
	var fileEntry *FileEntry
	for _, entry := range w.files {
		if entry.URL == "http://localhost:"+PORT+r.URL.Path {
			fileEntry = &entry
			break
		}
	}
	w.filesLock.Unlock()

	// If the file entry was found, serve the file using its AbsolutePath
	if fileEntry != nil {
		// Print path to be served
		println("Serving file: " + fileEntry.AbsolutePath)
		http.ServeFile(wr, r, fileEntry.AbsolutePath)
		return
	}

	// If no file is found, return a 404 Not Found error
	http.NotFound(wr, r)
}
