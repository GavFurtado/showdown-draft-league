package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockGameRepository struct {
	mock.Mock
}

func (m *MockGameRepository) CreateGame(game *models.Game) (models.Game, error) {
	args := m.Called(game)
	var result models.Game
	if args.Get(0) != nil {
		result = args.Get(0).(models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameRepository) GetGameByID(id uuid.UUID) (models.Game, error) {
	args := m.Called(id)
	var result models.Game
	if args.Get(0) != nil {
		result = args.Get(0).(models.Game)
	}
	return result, args.Error(1)
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

func (m *MockGameRepository) GetPendingGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
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

func (m *MockGameRepository) UpdateGameScore(gameID uuid.UUID, player1Wins, player2Wins int) error {
	args := m.Called(gameID, player1Wins, player2Wins)
	return args.Error(0)
}

func (m *MockGameRepository) ReportGameResult(gameID, winnerID, loserID, reporterID uuid.UUID, player1Wins, player2Wins int, replayLinks []string) error {
	args := m.Called(gameID, winnerID, loserID, reporterID, player1Wins, player2Wins, replayLinks)
	return args.Error(0)
}

func (m *MockGameRepository) DisputeGame(gameID, reporterID uuid.UUID) error {
	args := m.Called(gameID, reporterID)
	return args.Error(0)
}

func (m *MockGameRepository) ResolveDisputedGame(gameID, winnerID, loserID uuid.UUID, player1Wins, player2Wins int, replayLinks []string) error {
	args := m.Called(gameID, winnerID, loserID, player1Wins, player2Wins, replayLinks)
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

func (m *MockGameRepository) GetPlayerRecordInLeague(playerID, leagueID uuid.UUID) (int64, int64, error) {
	args := m.Called(playerID, leagueID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockGameRepository) GetCurrentRoundNumber(leagueID uuid.UUID) (int, error) {
	args := m.Called(leagueID)
	return args.Int(0), args.Error(1)
}

func (m *MockGameRepository) CreateGamesForWeek(games []models.Game) error {
	args := m.Called(games)
	return args.Error(0)
}

func (m *MockGameRepository) UpdatePlayerRecordsAfterGame(winnerID, loserID uuid.UUID, winnerNewWins, winnerNewLosses, loserNewWins, loserNewLosses int) error {
	args := m.Called(winnerID, loserID, winnerNewWins, winnerNewLosses, loserNewWins, loserNewLosses)
	return args.Error(0)
}

func (m *MockGameRepository) DeleteGame(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGameRepository) GetPendingGamesByPlayer(playerID uuid.UUID) ([]models.Game, error) {
	args := m.Called(playerID)
	var result []models.Game
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Game)
	}
	return result, args.Error(1)
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

func (m *MockGameRepository) CreateGames(games []*models.Game) error {
	args := m.Called(games)
	return args.Error(0)
}
