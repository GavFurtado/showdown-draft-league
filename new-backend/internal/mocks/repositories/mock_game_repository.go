package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
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

func (m *MockGameRepository) GetGameByID(id uuid.UUID) (models.Game, error) {
	args := m.Called(id)
	return args.Get(0).(models.Game), args.Error(1)
}

func (m *MockGameRepository) GetGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	args := m.Called(leagueID)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) GetGamesByPlayer(playerID uuid.UUID) ([]models.Game, error) {
	args := m.Called(playerID)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) GetGamesByLeagueAndRound(leagueID uuid.UUID, roundNumber int) ([]models.Game, error) {
	args := m.Called(leagueID, roundNumber)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) DisputeGame(gameID uuid.UUID, reporterID uuid.UUID) error {
	args := m.Called(gameID, reporterID)
	return args.Error(0)
}

func (m *MockGameRepository) GetHeadToHeadRecord(player1ID, player2ID uuid.UUID) ([]models.Game, error) {
	args := m.Called(player1ID, player2ID)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) GetPlayerRecordInLeague(playerID, leagueID uuid.UUID) (wins, losses int64, err error) {
	args := m.Called(playerID, leagueID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockGameRepository) CreateGames(games []*models.Game) error {
	args := m.Called(games)
	return args.Error(0)
}

func (m *MockGameRepository) DeleteGame(gameID uuid.UUID) error {
	args := m.Called(gameID)
	return args.Error(0)
}

func (m *MockGameRepository) GetScheduledGamesByPlayer(playerID uuid.UUID) ([]models.Game, error) {
	args := m.Called(playerID)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) GetScheduledGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	args := m.Called(leagueID)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) GetCompletedGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	args := m.Called(leagueID)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) GetDisputedGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	args := m.Called(leagueID)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) HasGames(leagueID uuid.UUID, gameType enums.GameType) (bool, error) {
	args := m.Called(leagueID, gameType)
	return args.Bool(0), args.Error(1)
}

func (m *MockGameRepository) UpdateGameReport(gameID uuid.UUID, loserID uuid.UUID, dto *common.ReportGameDTO) error {
	args := m.Called(gameID, loserID, dto)
	return args.Error(0)
}

func (m *MockGameRepository) FinalizeGameAndUpdateStats(game *models.Game, loserID uuid.UUID, dto *common.FinalizeGameDTO) error {
	args := m.Called(game, loserID, dto)
	return args.Error(0)
}
