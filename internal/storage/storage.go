package storage

import (
	"fmt"
	"log"
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

func New() *Storage {
	directory := os.Getenv("PHOTO_STORAGE_DIRECTORY")
	if directory == "" {
		log.Fatal("no storage directory was provided")
	}

	directory, err := filepath.Abs(directory)
	if err != nil {
		log.Fatalf("filepath.Abd: %v\n", err)
	}

	_, err = os.Stat(directory)
	if err != nil {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			log.Fatalf("os.MkdirAll: %v\n", err)
		}
	}

	_, err = os.Stat(filepath.Join(directory, "thumbnails"))
	if err != nil {
		err = os.MkdirAll(filepath.Join(directory, "thumbnails"), 0755)
		if err != nil {
			log.Fatalf("os.MkdirAll: %v\n", err)
		}
	}

	return &Storage{
		directory,
	}
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
