package database

import (
	"strconv"

	"github.com/grqphical/f-stop/internal/auth"
	"github.com/grqphical/f-stop/internal/models"
)

var userIdCounter int = 0

type MockDatabase struct {
	store map[string]any
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		store: make(map[string]any),
	}
}

func (m *MockDatabase) Close() {}

func (m *MockDatabase) CreateUser(username string, email string, password string) (models.User, error) {
	var user models.User
	user.Email = email
	user.ID = userIdCounter
	userIdCounter++

	hash, err := auth.GenerateHashFromPassword(password)
	if err != nil {
		return models.User{}, err
	}

	user.PasswordHash = hash

	key := strconv.Itoa(user.ID)
	m.store[key] = user

	return user, nil
}
