package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockSpectatorRepository struct {
	mock.Mock
}

func (m *MockSpectatorRepository) Create(spectator *models.Spectator) (*models.Spectator, error) {
	args := m.Called(spectator)
	var result *models.Spectator
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Spectator)
	}
	return result, args.Error(1)
}

func (m *MockSpectatorRepository) GetByID(id uuid.UUID) (*models.Spectator, error) {
	args := m.Called(id)
	var result *models.Spectator
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Spectator)
	}
	return result, args.Error(1)
}

func (m *MockSpectatorRepository) GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.Spectator, error) {
	args := m.Called(userID, leagueID)
	var result *models.Spectator
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Spectator)
	}
	return result, args.Error(1)
}

func (m *MockSpectatorRepository) GetByLeague(leagueID uuid.UUID) ([]models.Spectator, error) {
	args := m.Called(leagueID)
	var result []models.Spectator
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Spectator)
	}
	return result, args.Error(1)
}

func (m *MockSpectatorRepository) GetByUser(userID uuid.UUID) ([]models.Spectator, error) {
	args := m.Called(userID)
	var result []models.Spectator
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Spectator)
	}
	return result, args.Error(1)
}

func (m *MockSpectatorRepository) Update(spectator *models.Spectator) (*models.Spectator, error) {
	args := m.Called(spectator)
	var result *models.Spectator
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Spectator)
	}
	return result, args.Error(1)
}

func (m *MockSpectatorRepository) UpdateRole(spectatorID uuid.UUID, role rbac.MemberRole) error {
	args := m.Called(spectatorID, role)
	return args.Error(0)
}

func (m *MockSpectatorRepository) Delete(spectatorID uuid.UUID) error {
	args := m.Called(spectatorID)
	return args.Error(0)
}
