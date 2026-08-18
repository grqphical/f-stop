package storage

import (
	"log"
	"os"
	"path/filepath"
)

type MockStorage struct {
	storageLocation string
}

func NewMockInterface() *MockStorage {
	path, err := os.MkdirTemp("", "f-stop-tests")
	if err != nil {
		log.Fatal(err)
	}

	return &MockStorage{
		storageLocation: path,
	}
}

func (m *MockStorage) GetStorageLocation() string {
	return m.storageLocation
}

func (m *MockStorage) StoreItem(filename string) string {
	return filepath.Join(m.storageLocation, filename)
}
