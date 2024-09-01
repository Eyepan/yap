package cacher

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// FSCache handles caching for fetched data
type FSCache struct {
	CacheDir string
	mu       sync.Mutex
}

func (c *FSCache) Fetch(url string, cacheFilePath string, authToken string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create the full path including directory structure
	fullPath := filepath.Join(c.CacheDir, cacheFilePath)

	// Ensure the directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Check if the file exists
	if _, err := os.Stat(fullPath); err == nil {
		return os.ReadFile(fullPath)
	}

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set the Authorization header if authToken is provided
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	// Perform the HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Write the data to the file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return nil, err
	}

	return data, nil
}
