package storage

import (
	"log"
	"os"
	"path/filepath"

	_ "github.com/joho/godotenv/autoload"
)

type StorageInterface struct {
	directory string
}

func New() *StorageInterface {
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
		err = os.MkdirAll(directory, 0644)
		if err != nil {
			log.Fatalf("os.MkdirAll: %v\n", err)
		}
	}

	return &StorageInterface{
		directory,
	}
}

func (s *StorageInterface) GetStorageLocation() string {
	return s.directory
}

// Given a filename, this function returns the fullpath where the file should be written to
func (s *StorageInterface) StoreItem(filename string) string {
	return filepath.Join(s.directory, filename)
}
