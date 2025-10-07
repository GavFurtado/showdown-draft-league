package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockLeagueRepository is a mock implementation of repositories.LeagueRepository
type MockLeagueRepository struct {
	mock.Mock
}

func (m *MockLeagueRepository) CreateLeague(league *models.League) (*models.League, error) {
	args := m.Called(league)
	var result *models.League
	if args.Get(0) != nil {
		result = args.Get(0).(*models.League)
	}
	return result, args.Error(1)
}
func (m *MockLeagueRepository) GetLeagueByID(leagueID uuid.UUID) (*models.League, error) {
	args := m.Called(leagueID)
	var result *models.League
	if args.Get(0) != nil {
		result = args.Get(0).(*models.League)
	}
	return result, args.Error(1)
}
func (m *MockLeagueRepository) GetLeaguesCountWhereOwner(ownerID uuid.UUID) (int64, error) {
	args := m.Called(ownerID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockLeagueRepository) UpdateLeague(league *models.League) (*models.League, error) {
	args := m.Called(league)
	var result *models.League
	if args.Get(0) != nil {
		result = args.Get(0).(*models.League)
	}
	return result, args.Error(1)
}
func (m *MockLeagueRepository) GetLeaguesByOwner(ownerID uuid.UUID) ([]models.League, error) {
	args := m.Called(ownerID)
	return args.Get(0).([]models.League), args.Error(1)
}
func (m *MockLeagueRepository) GetLeaguesByUser(userID uuid.UUID) ([]models.League, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.League), args.Error(1)
}
func (m *MockLeagueRepository) IsUserOwner(userID, leagueID uuid.UUID) (bool, error) {
	args := m.Called(userID, leagueID)
	return args.Bool(0), args.Error(1)
}
func (m *MockLeagueRepository) DeleteLeague(leagueID uuid.UUID) error {
	args := m.Called(leagueID)
	return args.Error(0)
}
func (m *MockLeagueRepository) GetLeagueWithFullDetails(id uuid.UUID) (*models.League, error) {
	args := m.Called(id)
	var result *models.League
	if args.Get(0) != nil {
		result = args.Get(0).(*models.League)
	}
	return result, args.Error(1)
}
func (m *MockLeagueRepository) IsUserPlayerInLeague(userID, leagueID uuid.UUID) (bool, error) {
	args := m.Called(userID, leagueID)
	return args.Bool(0), args.Error(1)
}
func (m *MockLeagueRepository) GetLeagueStatus(leagueID uuid.UUID) (enums.LeagueStatus, error) {
	args := m.Called(leagueID)
	return args.Get(0).(enums.LeagueStatus), args.Error(1)
}
func (m *MockLeagueRepository) GetLeagueStatus(leagueID uuid.UUID) (enums.LeagueStatus, error) {
	args := m.Called(leagueID)
	return args.Get(0).(enums.LeagueStatus), args.Error(1)
}
