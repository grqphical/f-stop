package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type MockStorage struct {
	storageLocation   string
	thumbnailLocation string
}

func NewMockInterface() *MockStorage {
	storagePath, err := os.MkdirTemp("", "f-stop-tests")
	if err != nil {
		log.Fatal(err)
	}

	thumbnailPath, err := os.MkdirTemp("", "f-stop-tests-thumbnails")
	if err != nil {
		log.Fatal(err)
	}

	return &MockStorage{
		storageLocation:   storagePath,
		thumbnailLocation: thumbnailPath,
	}
}

func (m *MockStorage) GetStorageLocation() string {
	return m.storageLocation
}

func (m *MockStorage) StoreItem(filename string) string {
	return filepath.Join(m.storageLocation, filename)
}

func (m *MockStorage) StoreThumbnail(photoId string) string {
	return filepath.Join(m.thumbnailLocation, fmt.Sprintf("%s.jpg", photoId))
}
