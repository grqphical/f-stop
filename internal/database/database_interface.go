package database

import (
	"time"

	"github.com/grqphical/f-stop/internal/models"
)

// Generic interface that allows a mock database to be used for integration tests
type DBInterface interface {
	Close()

	CreateUser(string, string, string) (models.User, error)
	GetUserByEmail(string) (models.User, error)
	GetUserByID(int) (models.User, error)

	CreatePhotoMetadata(int64, string, int) (string, error)
	UpdatePhotoMetadataFilePath(string, string) error
	SetPhotoThumbnailPath(string, string) error
	SetPhotoThumbnailJobId(string, int) error
	SetPhotoEXIFData(string, *float64, *float64, *time.Time, *string) error
	GetPhotoMetadataFromID(string) (models.PhotoMetadata, error)
	GetUserPhotos(int) ([]models.PhotoMetadata, error)
	DeletePhoto(string) error

	EnqueueJob(payload models.JobPayload) (int, error)
	DequeueJob(batch_size int, max_retries int) (models.Job, error)
	AcknowledgeSuccess(job_id int) error
	AcknowledgeFailure(job_id int) error
	GetJob(job_id int) (models.Job, error)

	CreateTag(name string, ownerId int) (int, error)
	GetTagByName(name string) (models.Tag, error)
	GetTagByID(id int) (models.Tag, error)
	GetUserTags(ownerId int) ([]models.Tag, error)
	RenameTag(tagId int, newName string)
	DeleteTag(name string) error

	AssignPhotoTag(tag string, photoId string) error
	RemovePhotoTag(tag string, photoId string) error
	GetPhotoTags(photoId string) ([]models.Tag, error)
}
