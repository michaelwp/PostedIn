package storage

import (
	"encoding/json"
	"os"

	"PostedIn/internal/models"
)

type JSONStorage struct {
	filename string
}

func NewJSONStorage(filename string) *JSONStorage {
	return &JSONStorage{
		filename: filename,
	}
}

func (js *JSONStorage) LoadPosts() ([]models.Post, error) {
	data, err := os.ReadFile(js.filename)
	if err != nil {
		return []models.Post{}, nil // File doesn't exist yet
	}

	var posts []models.Post
	err = json.Unmarshal(data, &posts)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (js *JSONStorage) SavePosts(posts []models.Post) error {
	data, err := json.MarshalIndent(posts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(js.filename, data, 0644)
}