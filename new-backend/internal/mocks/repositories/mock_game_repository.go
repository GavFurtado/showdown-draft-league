package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockGameRepository struct {
	mock.Mock
}

func (m *MockGameRepository) CreateGame(game *models.Game) (*models.Game, error) {
	args := m.Called(game)
	var result *models.Game
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) GetGameByID(id uuid.UUID) (*models.Game, error) {
	args := m.Called(id)
	var result *models.Game
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) GetGamesByLeagueID(leagueID uuid.UUID) ([]models.Game, error) {
	args := m.Called(leagueID)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) GetGamesByPlayerID(playerID uuid.UUID) ([]models.Game, error) {
	args := m.Called(playerID)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) UpdateGame(game *models.Game) (*models.Game, error) {
	args := m.Called(game)
	var result *models.Game
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) DeleteGame(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
