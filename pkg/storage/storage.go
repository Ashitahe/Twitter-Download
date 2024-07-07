package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

// URLStorage defines the interface for URL storage operations
type URLStorage interface {
	AddURL(url string)
	URLExists(url string) bool
	RemoveURL(url string) error
	SaveToFile() error
	LoadFromFile() error
}

// URLStore holds the downloaded URLs
type URLStore struct {
	URLStoreFilePath string
	URLs map[string]bool
}

// NewURLStore creates a new URLStore
func NewURLStore(storeFileName string) *URLStore {
	return &URLStore{ URLStoreFilePath: storeFileName, URLs: make(map[string]bool)}
}

// AddURL adds a URL to the store
func (s *URLStore) AddURL(url string) {
	s.URLs[url] = true
}

// URLExists checks if a URL exists in the store
func (s *URLStore) URLExists(url string) bool {
	_, exists := s.URLs[url]
	return exists
}

// RemoveURL removes a URL from the store
func (s *URLStore) RemoveURL(url string) error {
	if _, exists := s.URLs[url]; !exists {
		return fmt.Errorf("URL not found: %s", url)
	}
	delete(s.URLs, url)
	return nil
}

// SaveToFile saves the URL store to a file
func (s *URLStore) SaveToFile() error {
	data, err := json.Marshal(s.URLs)
	if err != nil {
		return err
	}
	return os.WriteFile(s.URLStoreFilePath, data, 0644)
}

// LoadFromFile loads the URL store from a file
func (s *URLStore) LoadFromFile() error {
	data, err := os.ReadFile(s.URLStoreFilePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.URLs)
}
