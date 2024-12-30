package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

// URLMap stores our path mappings with thread-safe access
type URLMap struct {
	paths map[string]string
	mu    sync.RWMutex
}

// MapHandler will map paths to their corresponding URLs and redirect when applicable
func MapHandler(pathMap *URLMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the path from the request
		path := r.URL.Path

		// Check if we have a redirect for this path
		pathMap.mu.RLock()
		target, exists := pathMap.paths[path]
		pathMap.mu.RUnlock()

		if exists {
			// If we have a mapping, redirect to the target URL
			http.Redirect(w, r, target, http.StatusFound)
			return
		}

		// If we don't have a redirect, show a 404 page
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Page not found. Path: %s", path)
	}
}

// AddPath adds a new path mapping
func (u *URLMap) AddPath(path, url string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.paths[path] = url
}

func main() {
	// Initialize our URL map
	pathMap := &URLMap{
		paths: make(map[string]string),
	}

	// Add some example path mappings
	pathMap.AddPath("/github", "https://github.com")
	pathMap.AddPath("/google", "https://google.com")
	pathMap.AddPath("/stackoverflow", "https://stackoverflow.com")

	// Create a handler using our map
	handler := MapHandler(pathMap)

	// Add routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	// Add an endpoint to create new mappings
	mux.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := r.FormValue("path")
		url := r.FormValue("url")

		if path == "" || url == "" {
			http.Error(w, "Path and URL are required", http.StatusBadRequest)
			return
		}

		// Add the new mapping
		pathMap.AddPath(path, url)
		fmt.Fprintf(w, "Successfully mapped %s to %s", path, url)
	})

	// Start the server
	fmt.Println("Starting server on :8080")
	fmt.Println("Example mappings:")
	fmt.Println("  http://localhost:8080/github -> https://github.com")
	fmt.Println("  http://localhost:8080/google -> https://google.com")
	fmt.Println("  http://localhost:8080/stackoverflow -> https://stackoverflow.com")
	fmt.Println("\nTo add new mappings, use POST /add with path and url parameters")

	log.Fatal(http.ListenAndServe(":8080", mux))
}
