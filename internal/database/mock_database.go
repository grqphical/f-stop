package database

import (
	"strconv"

	"github.com/grqphical/f-stop/internal/auth"
	"github.com/grqphical/f-stop/internal/models"
)

type MockDatabase struct {
	users         map[string]models.User
	userIdCounter int
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		users:         make(map[string]models.User),
		userIdCounter: 0,
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

func (m *MockDatabase) CreatePhotoMetadata(arg1 int64, arg2 string, arg3 int) (string, error) {
	return "", nil
}

func (m *MockDatabase) UpdatePhotoMetadataFilePath(id, filePath string) error {
	return nil
}

func (m *MockDatabase) GetPhotoMetadataFromID(id string) (models.PhotoMetadata, error) {
	return models.PhotoMetadata{}, nil
}

func (m *MockDatabase) DeletePhoto(id string) error {
	return nil
}
