// main.go
package main

import (
	"flag"
	"path/filepath"
	"strings"
)

var PORT string

func main() {
	directoryFlag := flag.String("directory", "", "Directories to watch")
	extensionsFlag := flag.String("extensions", "", "File extensions to watch")
	portFlag := flag.String("port", "9191", "Port to serve on")
	flag.Parse()

  // Re-assign to global port variable
  PORT = *portFlag

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
