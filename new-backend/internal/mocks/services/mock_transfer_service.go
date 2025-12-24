package mock_services

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockTransferService is a mock implementation of services.TransferService
type MockTransferService struct {
	mock.Mock
}

func (m *MockTransferService) StartTransferPeriod(leagueID uuid.UUID) error {
	args := m.Called(leagueID)
	return args.Error(0)
}

func (m *MockTransferService) EndTransferPeriod(leagueID uuid.UUID) error {
	args := m.Called(leagueID)
	return args.Error(0)
}

func (m *MockTransferService) DropPokemon(currentUser *models.User, leagueID, draftedPokemonID uuid.UUID) error {
	args := m.Called(currentUser, leagueID, draftedPokemonID)
	return args.Error(0)
}

func (m *MockTransferService) PickupFreeAgent(currentUser *models.User, leagueID, leaguePokemonID uuid.UUID) error {
	args := m.Called(currentUser, leagueID, leaguePokemonID)
	return args.Error(0)
}

func (m *MockTransferService) SetSchedulerService(schedulerService services.SchedulerService) {
	m.Called(schedulerService)
}

func (m *MockTransferService) SetLeagueService(leagueService services.LeagueService) {
	m.Called(leagueService)
}
