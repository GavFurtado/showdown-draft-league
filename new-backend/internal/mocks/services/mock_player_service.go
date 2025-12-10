package mock_services

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockPlayerService is a mock implementation of services.PlayerService
type MockPlayerService struct {
	mock.Mock
}

func (m *MockPlayerService) CreatePlayerHandler(input *common.PlayerCreateRequest) (*models.Player, error) {
	args := m.Called(input)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerService) GetPlayerByIDHandler(playerID uuid.UUID, currentUser *models.User) (*models.Player, error) {
	args := m.Called(playerID, currentUser)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerService) GetPlayerByUserIDAndLeagueID(userID uuid.UUID, leagueID uuid.UUID) (*models.Player, error) {
	args := m.Called(userID, leagueID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerService) GetPlayersByLeagueHandler(leagueID, userID uuid.UUID) ([]models.Player, error) {
	args := m.Called(leagueID, userID)
	return args.Get(0).([]models.Player), args.Error(1)
}

func (m *MockPlayerService) GetPlayersByUserHandler(userID, currentUserID uuid.UUID) ([]models.Player, error) {
	args := m.Called(userID, currentUserID)
	return args.Get(0).([]models.Player), args.Error(1)
}

func (m *MockPlayerService) GetPlayerWithFullRosterHandler(playerID, currentUserID uuid.UUID) (*models.Player, error) {
	args := m.Called(playerID, currentUserID)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerService) GetPlayerRosterByWeek(playerID uuid.UUID, weekNumber int) ([]models.DraftedPokemon, error) {
	args := m.Called(playerID, weekNumber)
	return args.Get(0).([]models.DraftedPokemon), args.Error(1)
}

func (m *MockPlayerService) UpdatePlayerProfile(currentUser *models.User, playerID uuid.UUID, inLeagueName *string, teamName *string) (*models.Player, error) {
	args := m.Called(currentUser, playerID, inLeagueName, teamName)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerService) UpdatePlayerDraftPoints(currentUser *models.User, playerID uuid.UUID, draftPoints *int) (*models.Player, error) {
	args := m.Called(currentUser, playerID, draftPoints)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerService) UpdatePlayerRecord(currentUser *models.User, playerID uuid.UUID, wins int, losses int) (*models.Player, error) {
	args := m.Called(currentUser, playerID, wins, losses)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerService) UpdatePlayerDraftPosition(currentUser *models.User, playerID uuid.UUID, draftPosition int) (*models.Player, error) {
	args := m.Called(currentUser, playerID, draftPosition)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}

func (m *MockPlayerService) UpdatePlayerRole(currentUserID, playerID uuid.UUID, newPlayerRole rbac.PlayerRole) (*models.Player, error) {
	args := m.Called(currentUserID, playerID, newPlayerRole)
	var result *models.Player
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Player)
	}
	return result, args.Error(1)
}
