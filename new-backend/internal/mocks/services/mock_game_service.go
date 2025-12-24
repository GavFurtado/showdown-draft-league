package mock_services

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockGameService is a mock implementation of services.GameService
type MockGameService struct {
	mock.Mock
}

func (m *MockGameService) GetGameByID(ID uuid.UUID) (*models.Game, error) {
	args := m.Called(ID)
	var result *models.Game
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Game)
	}
	return result, args.Error(1)
}

func (m *MockGameService) GetGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.Game), args.Error(1)
}

func (m *MockGameService) GetGamesByPlayer(playerID uuid.UUID) ([]models.Game, error) {
	args := m.Called(playerID)
	return args.Get(0).([]models.Game), args.Error(1)
}

func (m *MockGameService) GenerateRegularSeasonGames(leagueID uuid.UUID) error {
	args := m.Called(leagueID)
	return args.Error(0)
}

func (m *MockGameService) GeneratePlayoffBracket(leagueID uuid.UUID) error {
	args := m.Called(leagueID)
	return args.Error(0)
}

func (m *MockGameService) ReportGameResult(gameID uuid.UUID, dto *common.ReportGameDTO) error {
	args := m.Called(gameID, dto)
	return args.Error(0)
}

func (m *MockGameService) FinalizeGameResult(gameID uuid.UUID, dto *common.FinalizeGameDTO) error {
	args := m.Called(gameID, dto)
	return args.Error(0)
}

func (m *MockGameService) SetLeagueService(leagueService services.LeagueService) {
	m.Called(leagueService)
}
