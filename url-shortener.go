package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	//"sync"
)

// URLMap stores our path mappings with thread-safe access
type URLMap struct {
	// paths map[string]string
	// mu    sync.RWMutex
}

// YAMLPathMap represents the structure of the YAML file.
type YAMLPathMap struct {
	Paths map[string]string `yaml:"paths"`
}

// MapHandler will map paths to their corresponding URLs and redirect when applicable
func MapHandler(pathsToUrls map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if dest, ok := pathsToUrls[path]; ok {
			http.Redirect(w, r, dest, http.StatusFound)
			return
		}
		http.NotFound(w, r) // Use http.NotFound for a standard 404 response
	}
}

// loadJSONFile reads a JSON file and returns a map of paths to URLs.
func loadJSONFile(jsonFile string) (map[string]string, error) {

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	var pathMap map[string]string
	if err := json.Unmarshal(data, &pathMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return pathMap, nil
}

func main() {
	const jsonFile = "urls.json"

	// Load path mappings from the JSON file
	pathsToUrls, err := loadJSONFile(jsonFile)

	if err != nil {
		// Handle error more gracefully
		if !os.IsNotExist(err) {
			log.Fatalf("Error loading JSON file: %v", err)
		}
		fmt.Println("urls.json not found, using default mappings.")
		pathsToUrls = map[string]string{
			"/github": "https://github.com",
			"/google": "https://google.com",
			// Add more default mappings here as needed
		}
	}

	// Create the default map handler
	mapHandler := MapHandler(pathsToUrls)

	// Create a new ServeMux
	mux := http.NewServeMux()
	mux.Handle("/", mapHandler) //Use mapHandler instead of MapHandler

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

		pathsToUrls[path] = url

		w.WriteHeader(http.StatusCreated) // Return 201 Created
		fmt.Fprintf(w, "Successfully mapped %s to %s\n", path, url)
	})

	mux.HandleFunc("/mappings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		jsonData, err := json.MarshalIndent(pathsToUrls, "", "    ")
		if err != nil {
			http.Error(w, "Error marshalling data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
