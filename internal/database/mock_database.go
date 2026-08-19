package database

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/grqphical/f-stop/internal/auth"
	"github.com/grqphical/f-stop/internal/models"
)

type MockDatabase struct {
	users         map[string]models.User
	userIdCounter int
	photos        map[string]models.PhotoMetadata
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		users:         make(map[string]models.User),
		userIdCounter: 0,
		photos:        make(map[string]models.PhotoMetadata),
	}
}

func (m *MockDatabase) Close() {}

func (m *MockDatabase) CreateUser(username string, email string, password string) (models.User, error) {
	for _, user := range m.users {
		if user.Email == email || user.Username == username {
			return models.User{}, ErrUniqueConstraint
		}
	}

	var user models.User
	user.Username = username
	user.Email = email
	user.ID = m.userIdCounter
	m.userIdCounter++

	hash, err := auth.GenerateHashFromPassword(password)
	if err != nil {
		return models.User{}, err
	}

	user.PasswordHash = hash

	key := strconv.Itoa(user.ID)
	m.users[key] = user

	return user, nil
}

func (m *MockDatabase) GetUserByEmail(email string) (models.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}

	return models.User{}, ErrNotFound
}

func (m *MockDatabase) GetUserByID(id int) (models.User, error) {
	stringId := strconv.Itoa(id)
	user, exists := m.users[stringId]
	if !exists {
		return models.User{}, ErrNotFound
	}

	return user, nil
}

func (m *MockDatabase) CreatePhotoMetadata(size int64, mimeType string, ownerID int) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	m.photos[id.String()] = models.PhotoMetadata{
		ID:       id.String(),
		OwnerID:  ownerID,
		Size:     int(size),
		MimeType: mimeType,
		Uploaded: time.Now(),
	}

	return id.String(), nil
}

func (m *MockDatabase) UpdatePhotoMetadataFilePath(id, filePath string) error {
	metadata, exists := m.photos[id]
	if !exists {
		return ErrNotFound
	}
	metadata.Filepath = filePath
	metadata.Permalink = fmt.Sprintf("/storage/%s", filepath.Base(filePath))

	m.photos[id] = metadata
	return nil
}

func (m *MockDatabase) GetPhotoMetadataFromID(id string) (models.PhotoMetadata, error) {
	metadata, exists := m.photos[id]
	if !exists {
		return metadata, ErrNotFound
	}

	return metadata, nil
}

func (m *MockDatabase) GetUserPhotos(id int) ([]models.PhotoMetadata, error) {
	result := make([]models.PhotoMetadata, 0)

	for _, photo := range m.photos {
		if photo.OwnerID == id {
			result = append(result, photo)
		}
	}

	return result, nil
}

func (m *MockDatabase) DeletePhoto(id string) error {
	delete(m.photos, id)
	return nil
}

func (m *MockDatabase) EnqueueJob(payload models.JobPayload) error    { return nil }
func (m *MockDatabase) DequeueJob(batch_size int) (models.Job, error) { return models.Job{}, nil }
func (m *MockDatabase) AcknowledgeSuccess(job_id int) error           { return nil }
func (m *MockDatabase) AcknowledgeFailure(job_id int) error           { return nil }
