package database

import "github.com/grqphical/f-stop/internal/models"

// Generic interface that allows a mock database to be used for integration tests
type DBInterface interface {
	Close()

	CreateUser(string, string, string) (models.User, error)
	GetUserByEmail(string) (models.User, error)
	GetUserByID(int) (models.User, error)

	CreatePhotoMetadata(int64, string, int) (string, error)
	UpdatePhotoMetadataFilePath(string, string) error
	GetPhotoMetadataFromID(string) (models.PhotoMetadata, error)
	GetUserPhotos(int) ([]models.PhotoMetadata, error)
	DeletePhoto(string) error
}
