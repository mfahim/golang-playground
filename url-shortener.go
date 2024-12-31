package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

// URLMap stores our path mappings with thread-safe access
type URLMap struct {
	paths map[string]string
	mu    sync.RWMutex
}
type PathMap struct {
	Mappings map[string]string `json:"mappings"`
}
type URLMapper struct {
	pathMap PathMap
	mu      sync.RWMutex
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

// NewURLMapper creates a new URLMapper from JSON file
func NewURLMapper(jsonFile string) (*URLMapper, error) {
	mapper := &URLMapper{
		pathMap: PathMap{
			Mappings: make(map[string]string),
		},
	}

	// Check if JSON file exists
	if _, err := os.Stat(jsonFile); err == nil {
		// Read and parse JSON file
		data, err := os.Open(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}

		var jsonData map[string]interface{}
		decoder := json.NewDecoder(data)
		err = decoder.Decode(&jsonData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}
	}

	return mapper, nil
}

func main() {
	const jsonFile = "urls.json"

	// Initialize URL mapper from JSON
	mapper, err := NewURLMapper(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize our URL map
	pathMap := &URLMap{
		paths: make(map[string]string),
	}
	// Add some example path mappings
	pathMap.AddPath("/github", "https://github.com")
	pathMap.AddPath("/google", "https://google.com")
	pathMap.AddPath("/stackoverflow", "https://stackoverflow.com")
	pathMap.AddPath("/urlshort", "https://github.com/gophercises/urlshort")
	pathMap.AddPath("/urlshort-final", "https://github.com/gophercises/urlshort/tree/solution")

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
	// Handler to view all mappings
	mux.HandleFunc("/mappings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mapper.mu.RLock()
		data, err := json.MarshalIndent(mapper.pathMap, "", "    ")
		mapper.mu.RUnlock()

		if err != nil {
			http.Error(w, "Error retrieving mappings", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
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
