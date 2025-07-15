// Package storage provides data persistence functionality for LinkedIn posts using JSON file storage.
package storage

import (
	"encoding/json"
	"os"

	"PostedIn/internal/models"
)

// JSONStorage provides JSON file-based storage for LinkedIn posts.
type JSONStorage struct {
	filename string
}

// NewJSONStorage creates a new JSON storage instance with the specified filename.
func NewJSONStorage(filename string) *JSONStorage {
	return &JSONStorage{
		filename: filename,
	}
}

// LoadPosts loads all posts from the JSON storage file.
func (js *JSONStorage) LoadPosts() ([]models.Post, error) {
	data, err := os.ReadFile(js.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Post{}, nil // File doesn't exist yet, return empty slice
		}

		return nil, err // Return the actual error for other cases
	}

	var posts []models.Post

	err = json.Unmarshal(data, &posts)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// SavePosts saves all posts to the JSON storage file.
func (js *JSONStorage) SavePosts(posts []models.Post) error {
	data, err := json.MarshalIndent(posts, "", "  ")
	if err != nil {
		return err
	}

	const restrictedPerm = 0o600

	return os.WriteFile(js.filename, data, restrictedPerm)
}
