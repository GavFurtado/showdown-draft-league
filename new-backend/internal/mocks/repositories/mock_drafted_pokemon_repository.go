package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockDraftedPokemonRepository is a mock implementation of repositories.DraftedPokemonRepository
type MockDraftedPokemonRepository struct {
	mock.Mock
}

func (m *MockDraftedPokemonRepository) CreateDraftedPokemon(draftedPokemon *models.DraftedPokemon) (*models.DraftedPokemon, error) {
	args := m.Called(draftedPokemon)
	var result *models.DraftedPokemon
	if args.Get(0) != nil {
		result = args.Get(0).(*models.DraftedPokemon)
	}
	return result, args.Error(1)
}
func (m *MockDraftedPokemonRepository) GetDraftedPokemonByID(id uuid.UUID) (*models.DraftedPokemon, error) {
	args := m.Called(id)
	var result *models.DraftedPokemon
	if args.Get(0) != nil {
		result = args.Get(0).(*models.DraftedPokemon)
	}
	return result, args.Error(1)
}
func (m *MockDraftedPokemonRepository) GetDraftedPokemonByPlayer(playerID uuid.UUID) ([]models.DraftedPokemon, error) {
	args := m.Called(playerID)
	return args.Get(0).([]models.DraftedPokemon), args.Error(1)
}
func (m *MockDraftedPokemonRepository) GetDraftedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.DraftedPokemon), args.Error(1)
}
func (m *MockDraftedPokemonRepository) GetActiveDraftedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.DraftedPokemon), args.Error(1)
}
func (m *MockDraftedPokemonRepository) GetReleasedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.DraftedPokemon), args.Error(1)
}

func (m *MockDraftedPokemonRepository) IsPokemonDrafted(leagueID uuid.UUID, pokemonSpeciesID int64) (bool, error) {
	args := m.Called(leagueID, pokemonSpeciesID)
	return args.Bool(0), args.Error(1)
}
func (m *MockDraftedPokemonRepository) GetNextDraftPickNumber(leagueID uuid.UUID) (int, error) {
	args := m.Called(leagueID)
	return args.Int(0), args.Error(1)
}
func (m *MockDraftedPokemonRepository) ReleasePokemon(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockDraftedPokemonRepository) ReDraftPokemon(draftedPokemonID, newPlayerID uuid.UUID, newPickNumber int) error {
	args := m.Called(draftedPokemonID, newPlayerID, newPickNumber)
	return args.Error(0)
}
func (m *MockDraftedPokemonRepository) GetDraftedPokemonCountByPlayer(playerID uuid.UUID) (int64, error) {
	args := m.Called(playerID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockDraftedPokemonRepository) GetDraftHistory(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.DraftedPokemon), args.Error(1)
}
func (m *MockDraftedPokemonRepository) TradePokemon(draftedPokemonID, newPlayerID uuid.UUID) error {
	args := m.Called(draftedPokemonID, newPlayerID)
	return args.Error(0)
}
func (m *MockDraftedPokemonRepository) DeleteDraftedPokemon(draftedPokemonID uuid.UUID) error {
	args := m.Called(draftedPokemonID)
	return args.Error(0)
}

func (m *MockDraftedPokemonRepository) GetActiveDraftedPokemonCountByLeague(leagueID uuid.UUID) (int64, error) {
	args := m.Called(leagueID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDraftedPokemonRepository) DraftPokemonBatchTransaction(draftedPokemon []*models.DraftedPokemon, player *models.Player, leaguePokemonIDs []uuid.UUID, totalCost int) error {
	args := m.Called(draftedPokemon, player, leaguePokemonIDs, totalCost)
	return args.Error(0)
}

func (m *MockDraftedPokemonRepository) PickupFreeAgentTransaction(player *models.Player, newDraftedPokemon *models.DraftedPokemon, leaguePokemon *models.LeaguePokemon) error {
	args := m.Called(player, newDraftedPokemon, leaguePokemon)
	return args.Error(0)
}
