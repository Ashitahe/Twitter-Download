package store

import (
	"encoding/json"
	"fmt"
	"os"
)

// URLStorer defines the interface for URL storage operations
type URLStorer interface {
	AddURL(url string)
	URLExists(url string) bool
	RemoveURL(url string) error
	SaveToFile(filename string) error
	LoadFromFile(filename string) error
}

// URLStore holds the downloaded URLs
type URLStore struct {
	URLs map[string]bool
}

// NewURLStore creates a new URLStore
func NewURLStore() *URLStore {
	return &URLStore{URLs: make(map[string]bool)}
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
func (s *URLStore) SaveToFile(filename string) error {
	data, err := json.Marshal(s.URLs)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile loads the URL store from a file
func (s *URLStore) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.URLs)
}
