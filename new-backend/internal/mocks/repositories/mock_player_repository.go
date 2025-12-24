package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockPlayerRepository is a mock implementation of repositories.PlayerRepository
type MockPlayerRepository struct {
	mock.Mock
}

func (m *MockPlayerRepository) CreatePlayer(player *models.Player) (*models.Player, error) {
	args := m.Called(player)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}
func (m *MockPlayerRepository) GetPlayerByID(id uuid.UUID) (*models.Player, error) {
	args := m.Called(id)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}
func (m *MockPlayerRepository) GetPlayersByLeague(leagueID uuid.UUID) ([]models.Player, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.Player), args.Error(1)
}
func (m *MockPlayerRepository) GetPlayersByUser(userID uuid.UUID) ([]models.Player, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Player), args.Error(1)
}
func (m *MockPlayerRepository) GetPlayerWithFullRoster(playerID uuid.UUID) (*models.Player, error) {
	args := m.Called(playerID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}
func (m *MockPlayerRepository) UpdatePlayer(player *models.Player) (*models.Player, error) {
	args := m.Called(player)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}
func (m *MockPlayerRepository) UpdatePlayerDraftPoints(playerID uuid.UUID, newPoints int) error {
	args := m.Called(playerID, newPoints)
	return args.Error(0)
}
func (m *MockPlayerRepository) UpdatePlayerRecord(playerID uuid.UUID, wins, losses int) error {
	args := m.Called(playerID, wins, losses)
	return args.Error(0)
}
func (m *MockPlayerRepository) UpdatePlayerDraftPosition(playerID uuid.UUID, newPosition int) error {
	args := m.Called(playerID, newPosition)
	return args.Error(0)
}
func (m *MockPlayerRepository) UpdatePlayerRole(playerID uuid.UUID, playerRole rbac.PlayerRole) error {
	args := m.Called(playerID, playerRole)
	return args.Error(0)
}
func (m *MockPlayerRepository) FindPlayerByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error) {
	args := m.Called(userID, leagueID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}
func (m *MockPlayerRepository) FindPlayerByInLeagueNameAndLeagueID(inLeagueName string, leagueID uuid.UUID) (*models.Player, error) {
	args := m.Called(inLeagueName, leagueID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}
func (m *MockPlayerRepository) FindPlayerByTeamNameAndLeagueID(teamName string, leagueID uuid.UUID) (*models.Player, error) {
	args := m.Called(teamName, leagueID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}
func (m *MockPlayerRepository) GetPlayerCountByLeague(leagueID uuid.UUID) (int64, error) {
	args := m.Called(leagueID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockPlayerRepository) GetPlayerByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error) {
	args := m.Called(userID, leagueID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}
func (m *MockPlayerRepository) DeletePlayer(playerID uuid.UUID) error {
	args := m.Called(playerID)
	return args.Error(0)
}
func (m *MockPlayerRepository) IsUserInLeague(userID, leagueID uuid.UUID) (bool, error) {
	args := m.Called(userID, leagueID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayersByLeagueAndGroupNumber(leagueID uuid.UUID, groupNumber int) ([]models.Player, error) {
	args := m.Called(leagueID, groupNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Player), args.Error(1)
}
