package fetcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Cache structure
type Cache struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewCache initializes a new cache
func NewCache() *Cache {
	return &Cache{
		data: make(map[string][]byte),
	}
}

// Fetch function with caching, JSON conversion, and token-based authorization
func (c *Cache) Fetch(url string, token string, target interface{}) error {
	// Check if the URL is already in the cache
	c.mu.RLock()
	if cachedData, found := c.data[url]; found {
		c.mu.RUnlock()
		fmt.Println("CACHE ", url)
		// Unmarshal the cached JSON into the target
		return json.Unmarshal(cachedData, target)
	}
	c.mu.RUnlock()

	// If not in cache, perform the GET request with token-based authorization
	fmt.Println("GET", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to fetch data: " + resp.Status)
	}

	// Cache the response
	c.mu.Lock()
	c.data[url] = body
	c.mu.Unlock()

	// Unmarshal the JSON into the target
	return json.Unmarshal(body, target)
}
