package fetcher

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// FSCache manages caching for fetched data.
type FSCache struct {
	CacheDir string
	mu       sync.Mutex
}

// Fetch retrieves data from the cache or fetches it if not cached.
func (c *FSCache) Fetch(url, cacheFilePath, authToken string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullPath := filepath.Join(c.CacheDir, cacheFilePath)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	if data, err := os.ReadFile(fullPath); err == nil {
		return data, nil
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return nil, err
	}

	return data, nil
}
