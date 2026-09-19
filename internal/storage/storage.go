package storage

import (
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/joho/godotenv/autoload"
)

type StorageInterface interface {
	GetStorageLocation() string
	StoreItem(string) string
	StoreThumbnail(string) string
}

type Storage struct {
	directory string
}

func New() (*Storage, error) {
	directory := os.Getenv("PHOTO_STORAGE_DIRECTORY")
	if directory == "" {
		return nil, fmt.Errorf("no storage directory was provided")
	}

	directory, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("filepath.Abs: %w", err)
	}

	_, err = os.Stat(directory)
	if err != nil {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return nil, fmt.Errorf("os.MkdirAll: %w", err)
		}
	}

	_, err = os.Stat(filepath.Join(directory, "thumbnails"))
	if err != nil {
		err = os.MkdirAll(filepath.Join(directory, "thumbnails"), 0755)
		if err != nil {
			return nil, fmt.Errorf("os.MkdirAll: %w", err)
		}
	}

	return &Storage{
		directory,
	}, nil
}

func (s *Storage) GetStorageLocation() string {
	return s.directory
}

// Given a filename, this function returns the fullpath where the file should be written to
func (s *Storage) StoreItem(filename string) string {
	return filepath.Join(s.directory, filename)
}

func (s *Storage) StoreThumbnail(photoId string) string {
	return filepath.Join(s.directory, "thumbnails", fmt.Sprintf("%s.jpg", photoId))
}
