package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of repositories.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) (*models.User, error) {
	args := m.Called(user)
	var result *models.User
	if args.Get(0) != nil {
		result = args.Get(0).(*models.User)
	}
	return result, args.Error(1)
}
func (m *MockUserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	var result *models.User
	if args.Get(0) != nil {
		result = args.Get(0).(*models.User)
	}
	return result, args.Error(1)
}
func (m *MockUserRepository) GetUserByDiscordID(discordID string) (*models.User, error) {
	args := m.Called(discordID)
	var result *models.User
	if args.Get(0) != nil {
		result = args.Get(0).(*models.User)
	}
	return result, args.Error(1)
}
func (m *MockUserRepository) UpdateUser(user *models.User) (*models.User, error) {
	args := m.Called(user)
	var result *models.User
	if args.Get(0) != nil {
		result = args.Get(0).(*models.User)
	}
	return result, args.Error(1)
}
func (m *MockUserRepository) GetUserLeagues(userID uuid.UUID) ([]models.League, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.League), args.Error(1)
}
