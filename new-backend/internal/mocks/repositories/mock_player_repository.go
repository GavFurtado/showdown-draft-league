package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockPlayerRepository struct {
	mock.Mock
}

func (m *MockPlayerRepository) Create(player *models.Player) (*models.Player, error) {
	args := m.Called(player)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerRepository) GetByID(id uuid.UUID) (*models.Player, error) {
	args := m.Called(id)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerRepository) GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error) {
	args := m.Called(userID, leagueID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerRepository) GetByLeague(leagueID uuid.UUID) ([]models.Player, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.Player), args.Error(1)
}

func (m *MockPlayerRepository) GetByLeagueAndGroup(leagueID uuid.UUID, groupNumber int) ([]models.Player, error) {
	args := m.Called(leagueID, groupNumber)
	return args.Get(0).([]models.Player), args.Error(1)
}

func (m *MockPlayerRepository) GetByUser(userID uuid.UUID) ([]models.Player, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Player), args.Error(1)
}

func (m *MockPlayerRepository) Update(player *models.Player) (*models.Player, error) {
	args := m.Called(player)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerRepository) UpdateDraftPoints(playerID uuid.UUID, points int) error {
	args := m.Called(playerID, points)
	return args.Error(0)
}

func (m *MockPlayerRepository) UpdateRecord(playerID uuid.UUID, wins, losses int) error {
	args := m.Called(playerID, wins, losses)
	return args.Error(0)
}

func (m *MockPlayerRepository) UpdateDraftPosition(playerID uuid.UUID, position int) error {
	args := m.Called(playerID, position)
	return args.Error(0)
}

func (m *MockPlayerRepository) UpdateRole(playerID uuid.UUID, role rbac.MemberRole) error {
	args := m.Called(playerID, role)
	return args.Error(0)
}

func (m *MockPlayerRepository) GetCountByLeague(leagueID uuid.UUID) (int64, error) {
	args := m.Called(leagueID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPlayerRepository) Delete(playerID uuid.UUID) error {
	args := m.Called(playerID)
	return args.Error(0)
}

func (m *MockPlayerRepository) IsUserInLeague(userID, leagueID uuid.UUID) (bool, error) {
	args := m.Called(userID, leagueID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPlayerRepository) GetWithFullRoster(playerID uuid.UUID) (*models.Player, error) {
	args := m.Called(playerID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerRepository) FindByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error) {
	args := m.Called(userID, leagueID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerRepository) FindByInLeagueName(name string, leagueID uuid.UUID) (*models.Player, error) {
	args := m.Called(name, leagueID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerRepository) FindByTeamName(name string, leagueID uuid.UUID) (*models.Player, error) {
	args := m.Called(name, leagueID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}
