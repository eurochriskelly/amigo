// fileentry.go
package main

// FileEntry represents a file entry in the API response
type FileEntry struct {
	Label        string `json:"label"`
	Type         string `json:"type"`
	URL          string `json:"url"`
	AbsolutePath string `json:"absolutePath"`
}
