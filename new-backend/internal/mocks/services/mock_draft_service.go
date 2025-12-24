package mock_services

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockDraftService is a mock implementation of services.DraftService
type MockDraftService struct {
	mock.Mock
}

func (m *MockDraftService) GetDraftByID(draftID uuid.UUID) (*models.Draft, error) {
	args := m.Called(draftID)
	var result *models.Draft
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Draft)
	}
	return result, args.Error(1)
}

func (m *MockDraftService) GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error) {
	args := m.Called(leagueID)
	var result *models.Draft
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Draft)
	}
	return result, args.Error(1)
}

func (m *MockDraftService) StartDraft(leagueID uuid.UUID, TurnTimeLimit int) (*models.Draft, error) {
	args := m.Called(leagueID, TurnTimeLimit)
	var result *models.Draft
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Draft)
	}
	return result, args.Error(1)
}

func (m *MockDraftService) MakePick(currentUser *models.User, leagueID uuid.UUID, input *common.DraftMakePickDTO) error {
	args := m.Called(currentUser, leagueID, input)
	return args.Error(0)
}

func (m *MockDraftService) SkipTurn(currentUser *models.User, leagueID uuid.UUID) error {
	args := m.Called(currentUser, leagueID)
	return args.Error(0)
}

func (m *MockDraftService) AutoSkipTurn(playerID, leagueID uuid.UUID) error {
	args := m.Called(playerID, leagueID)
	return args.Error(0)
}

func (m *MockDraftService) SetSchedulerService(schedulerService services.SchedulerService) {
	m.Called(schedulerService)
}
