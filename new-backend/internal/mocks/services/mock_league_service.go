package mock_services

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockLeagueService is a mock implementation of services.LeagueService
type MockLeagueService struct {
	mock.Mock
}

func (m *MockLeagueService) CreateLeague(userID uuid.UUID, req *common.LeagueCreateRequestDTO) (*models.League, error) {
	args := m.Called(userID, req)
	var result *models.League
	if args.Get(0) != nil {
		result = args.Get(0).(*models.League)
	}
	return result, args.Error(1)
}

func (m *MockLeagueService) GetLeagueByIDForUser(userID, leagueID uuid.UUID) (*models.League, error) {
	args := m.Called(userID, leagueID)
	var result *models.League
	if args.Get(0) != nil {
		result = args.Get(0).(*models.League)
	}
	return result, args.Error(1)
}

func (m *MockLeagueService) GetLeaguesByCommissioner(userID uuid.UUID, currentUser *models.User) ([]models.League, error) {
	args := m.Called(userID, currentUser)
	return args.Get(0).([]models.League), args.Error(1)
}

func (m *MockLeagueService) GetLeaguesByUser(userID uuid.UUID, currentUser *models.User) ([]models.League, error) {
	args := m.Called(userID, currentUser)
	return args.Get(0).([]models.League), args.Error(1)
}

func (m *MockLeagueService) ProcessWeeklyTick(leagueID uuid.UUID) error {
	args := m.Called(leagueID)
	return args.Error(0)
}

func (m *MockLeagueService) SetSchedulerService(schedulerService services.SchedulerService) {
	m.Called(schedulerService)
}

func (m *MockLeagueService) SetGameService(gameService services.GameService) {
	m.Called(gameService)
}

func (m *MockLeagueService) SetTransferService(transferService services.TransferService) {
	m.Called(transferService)
}

func (m *MockLeagueService) StartRegularSeason(leagueID uuid.UUID) error {
	args := m.Called(leagueID)
	return args.Error(0)
}
