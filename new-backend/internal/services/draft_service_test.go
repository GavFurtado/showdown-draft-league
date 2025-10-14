package services_test

import (
	"testing"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type draftServiceMocks struct {
	draftRepo          *mock_repositories.MockDraftRepository
	leagueRepo         *mock_repositories.MockLeagueRepository
	playerRepo         *mock_repositories.MockPlayerRepository
	leaguePokemonRepo  *mock_repositories.MockLeaguePokemonRepository
	draftedPokemonRepo *mock_repositories.MockDraftedPokemonRepository
}

func setupDraftServiceTest() (services.DraftService, draftServiceMocks) {
	mocks := draftServiceMocks{
		draftRepo:          new(mock_repositories.MockDraftRepository),
		leagueRepo:         new(mock_repositories.MockLeagueRepository),
		playerRepo:         new(mock_repositories.MockPlayerRepository),
		leaguePokemonRepo:  new(mock_repositories.MockLeaguePokemonRepository),
		draftedPokemonRepo: new(mock_repositories.MockDraftedPokemonRepository),
	}

	// webhookService can be nil for these tests if not used
	service := services.NewDraftService(
		mocks.leagueRepo,
		mocks.leaguePokemonRepo,
		mocks.draftRepo,
		mocks.draftedPokemonRepo,
		mocks.playerRepo,
		nil,
	)

	return service, mocks
}

func TestDraftService_MakePick(t *testing.T) {
	service, mocks := setupDraftServiceTest()

	// --- Test Data ---
	leagueID := uuid.New()
	userID := uuid.New()
	playerID := uuid.New()
	leaguePokemonID := uuid.New()
	pokemonSpeciesID := int64(1)




	t.Run("Success - Make a valid pick", func(t *testing.T) {
		// --- Test Data (local to this test) ---
		localLeague := &models.League{
			ID:                  leagueID,
			Status:              enums.LeagueStatusDrafting,
			MinPokemonPerPlayer: 1,
			MaxPokemonPerPlayer: 12,
			Format:              models.LeagueFormat{IsSnakeRoundDraft: true},
		}
		localPlayer := &models.Player{ID: playerID, UserID: userID, LeagueID: leagueID, DraftPoints: 100}
		localDraft := &models.Draft{
			LeagueID:                    leagueID,
			Status:                      enums.DraftStatusOngoing,
			CurrentPickOnClock:          1,
			CurrentTurnPlayerID:         &playerID,
			PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks),
		}
		localInput := &common.DraftMakePickDTO{
			RequestedPickCount: 1,
			RequestedPicks: []common.RequestedPick{
				{LeaguePokemonID: leaguePokemonID, DraftPickNumber: 1},
			},
		}

		localCurrentUser := &models.User{ID: userID}
		localLeaguePokemon := &models.LeaguePokemon{
			ID:               leaguePokemonID,
			LeagueID:         leagueID,
			PokemonSpeciesID: pokemonSpeciesID,
			Cost:             func(i int) *int { return &i }(50),
			IsAvailable:      true,
		}

		// --- Mock Setup ---
		mocks.leagueRepo.On("GetLeagueByID", leagueID).Return(localLeague, nil).Once()
		mocks.draftRepo.On("GetDraftByLeagueID", leagueID).Return(localDraft, nil).Once()
		mocks.playerRepo.On("GetPlayerByUserAndLeague", userID, leagueID).Return(localPlayer, nil).Once()
		mocks.leaguePokemonRepo.On("GetLeaguePokemonByIDs", leagueID, []uuid.UUID{leaguePokemonID}).Return([]models.LeaguePokemon{*localLeaguePokemon}, nil).Once()
		mocks.playerRepo.On("GetPlayerCountByLeague", leagueID).Return(int64(1), nil).Once()
		mocks.draftedPokemonRepo.On("GetDraftedPokemonCountByPlayer", playerID).Return(int64(0), nil).Once()
		mocks.draftedPokemonRepo.On("DraftPokemonBatchTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		mocks.playerRepo.On("GetPlayersByLeague", leagueID).Return([]models.Player{*localPlayer}, nil).Once() // Simplified for this test
		mocks.draftRepo.On("UpdateDraft", mock.AnythingOfType("*models.Draft")).Return(nil).Once()
		mocks.draftedPokemonRepo.On("GetActiveDraftedPokemonCountByLeague", leagueID).Return(int64(1), nil).Once()

		// --- Call Service ---
		err := service.MakePick(localCurrentUser, leagueID, localInput)

		// --- Assertions ---
		assert.NoError(t, err)
		mocks.leagueRepo.AssertExpectations(t)
		mocks.draftRepo.AssertExpectations(t)
		mocks.playerRepo.AssertExpectations(t)
		mocks.leaguePokemonRepo.AssertExpectations(t)
		mocks.draftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Failure - Not player's turn", func(t *testing.T) {
		// --- Test Data (local to this test) ---
		otherPlayerID := uuid.New()
		localLeague := &models.League{
			ID:                  leagueID,
			Status:              enums.LeagueStatusDrafting,
			MinPokemonPerPlayer: 1,
			MaxPokemonPerPlayer: 12,
			Format:              models.LeagueFormat{IsSnakeRoundDraft: true},
		}
		localPlayer := &models.Player{ID: playerID, UserID: userID, LeagueID: leagueID, DraftPoints: 100}
		        localDraft := &models.Draft{
		            LeagueID:            leagueID,
		            Status:              enums.DraftStatusOngoing,
		            CurrentPickOnClock:  1,
		            CurrentTurnPlayerID: &otherPlayerID, // Not the current player's turn
		        }
		        localInput := &common.DraftMakePickDTO{
		            RequestedPickCount: 1,
		            RequestedPicks: []common.RequestedPick{
		                {LeaguePokemonID: leaguePokemonID, DraftPickNumber: 1},
		            },
		        }
		
		        		localCurrentUser := &models.User{ID: userID}
		        
		        		// --- Mock Setup ---
		        		mocks.leagueRepo.On("GetLeagueByID", leagueID).Return(localLeague, nil).Once()
		        		mocks.draftRepo.On("GetDraftByLeagueID", leagueID).Return(localDraft, nil).Once()
		        		mocks.playerRepo.On("GetPlayerByUserAndLeague", userID, leagueID).Return(localPlayer, nil).Once()
		        
		        		// --- Call Service ---
		        		err := service.MakePick(localCurrentUser, leagueID, localInput)		// --- Assertions ---
		assert.Error(t, err)
		assert.Equal(t, common.ErrUnauthorized, err)
		mocks.leagueRepo.AssertExpectations(t)
		mocks.draftRepo.AssertExpectations(t)
		mocks.playerRepo.AssertExpectations(t)
	})

	t.Run("Failure - League not in drafting status", func(t *testing.T) {
		// --- Test Data (local to this test) ---
		localLeague := &models.League{
			ID:     leagueID,
			Status: enums.LeagueStatusSetup, // Not drafting
		}
		localPlayer := &models.Player{ID: playerID, UserID: userID, LeagueID: leagueID, DraftPoints: 100}
		localDraft := &models.Draft{
			LeagueID:                    leagueID,
			Status:                      enums.DraftStatusOngoing,
			CurrentPickOnClock:          1,
			CurrentTurnPlayerID:         &playerID,
			PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks),
		}
		localInput := &common.DraftMakePickDTO{
			RequestedPickCount: 1,
			RequestedPicks: []common.RequestedPick{
				{LeaguePokemonID: leaguePokemonID, DraftPickNumber: 1},
			},
		}

		localCurrentUser := &models.User{ID: userID}

		// --- Mock Setup ---
		mocks.leagueRepo.On("GetLeagueByID", leagueID).Return(localLeague, nil).Once()
		mocks.draftRepo.On("GetDraftByLeagueID", leagueID).Return(localDraft, nil).Once()
		mocks.playerRepo.On("GetPlayerByUserAndLeague", userID, leagueID).Return(localPlayer, nil).Once()

		// --- Call Service ---
		err := service.MakePick(localCurrentUser, leagueID, localInput)

		// --- Assertions ---
		assert.Error(t, err)
		assert.Equal(t, common.ErrInvalidState, err)
		mocks.leagueRepo.AssertExpectations(t)
		mocks.draftRepo.AssertExpectations(t)
		mocks.playerRepo.AssertExpectations(t)
	})

	t.Run("Failure - Pokemon not available", func(t *testing.T) {
		// --- Test Data (local to this test) ---
		localLeague := &models.League{
			ID:                  leagueID,
			Status:              enums.LeagueStatusDrafting,
			MinPokemonPerPlayer: 1,
			MaxPokemonPerPlayer: 12,
			Format:              models.LeagueFormat{IsSnakeRoundDraft: true},
		}
		localPlayer := &models.Player{ID: playerID, UserID: userID, LeagueID: leagueID, DraftPoints: 100}
		localDraft := &models.Draft{
			LeagueID:                    leagueID,
			Status:                      enums.DraftStatusOngoing,
			CurrentPickOnClock:          1,
			CurrentTurnPlayerID:         &playerID,
			PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks),
		}
		localInput := &common.DraftMakePickDTO{
			RequestedPickCount: 1,
			RequestedPicks: []common.RequestedPick{
				{LeaguePokemonID: leaguePokemonID, DraftPickNumber: 1},
			},
		}
		unavailablePokemon := &models.LeaguePokemon{
			ID:          leaguePokemonID,
			IsAvailable: false, // Already drafted
		}

		localCurrentUser := &models.User{ID: userID}

		// --- Mock Setup ---
		mocks.leagueRepo.On("GetLeagueByID", leagueID).Return(localLeague, nil).Once()
		mocks.draftRepo.On("GetDraftByLeagueID", leagueID).Return(localDraft, nil).Once()
		mocks.playerRepo.On("GetPlayerByUserAndLeague", userID, leagueID).Return(localPlayer, nil).Once()
		mocks.leaguePokemonRepo.On("GetLeaguePokemonByIDs", leagueID, []uuid.UUID{leaguePokemonID}).Return([]models.LeaguePokemon{*unavailablePokemon}, nil).Once()

		// --- Call Service ---
		err := service.MakePick(localCurrentUser, leagueID, localInput)

		// --- Assertions ---
		assert.Error(t, err)
		assert.Equal(t, common.ErrConflict, err)
		mocks.leagueRepo.AssertExpectations(t)
		mocks.draftRepo.AssertExpectations(t)
		mocks.playerRepo.AssertExpectations(t)
		mocks.leaguePokemonRepo.AssertExpectations(t)
	})

	t.Run("Failure - Insufficient draft points", func(t *testing.T) {
		// --- Test Data (local to this test) ---
		localLeague := &models.League{
			ID:                  leagueID,
			Status:              enums.LeagueStatusDrafting,
			MinPokemonPerPlayer: 1,
			MaxPokemonPerPlayer: 12,
			Format:              models.LeagueFormat{IsSnakeRoundDraft: true},
		}
		localPlayer := &models.Player{ID: playerID, UserID: userID, LeagueID: leagueID, DraftPoints: 20} // Not enough points
		localDraft := &models.Draft{
			LeagueID:                    leagueID,
			Status:                      enums.DraftStatusOngoing,
			CurrentPickOnClock:          1,
			CurrentTurnPlayerID:         &playerID,
			PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks),
		}
		localInput := &common.DraftMakePickDTO{
			RequestedPickCount: 1,
			RequestedPicks: []common.RequestedPick{
				{LeaguePokemonID: leaguePokemonID, DraftPickNumber: 1},
			},
		}

		localCurrentUser := &models.User{ID: userID}
		localLeaguePokemon := &models.LeaguePokemon{
			ID:               leaguePokemonID,
			LeagueID:         leagueID,
			PokemonSpeciesID: pokemonSpeciesID,
			Cost:             func(i int) *int { return &i }(50),
			IsAvailable:      true,
		}

		// --- Mock Setup ---
		mocks.leagueRepo.On("GetLeagueByID", leagueID).Return(localLeague, nil).Once()
		mocks.draftRepo.On("GetDraftByLeagueID", leagueID).Return(localDraft, nil).Once()
		mocks.playerRepo.On("GetPlayerByUserAndLeague", userID, leagueID).Return(localPlayer, nil).Once()
		mocks.leaguePokemonRepo.On("GetLeaguePokemonByIDs", leagueID, []uuid.UUID{leaguePokemonID}).Return([]models.LeaguePokemon{*localLeaguePokemon}, nil).Once()
		mocks.playerRepo.On("GetPlayerCountByLeague", leagueID).Return(int64(8), nil).Once()
		mocks.draftedPokemonRepo.On("GetDraftedPokemonCountByPlayer", playerID).Return(int64(0), nil).Once()

		// --- Call Service ---
		err := service.MakePick(localCurrentUser, leagueID, localInput)

		// --- Assertions ---
		assert.Error(t, err)
		assert.Equal(t, common.ErrInsufficientDraftPoints, err)
		mocks.leagueRepo.AssertExpectations(t)
		mocks.draftRepo.AssertExpectations(t)
		mocks.playerRepo.AssertExpectations(t)
		mocks.leaguePokemonRepo.AssertExpectations(t)
		mocks.draftedPokemonRepo.AssertExpectations(t)
	})
}
