package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mock_repos "github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
)

func TestGameService_GeneratePlayoffBracket_Correct(t *testing.T) {
	// ARRANGE
	// Instantiate the EXISTING mocks from the mock_repositories package
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockPlayerRepo := new(mock_repos.MockPlayerRepository)
	mockGameRepo := new(mock_repos.MockGameRepository)

	leagueID := uuid.New()

	playerA := models.Player{ID: uuid.New(), Wins: 10, Losses: 0} // Seed 1
	playerB := models.Player{ID: uuid.New(), Wins: 8, Losses: 2}  // Seed 2
	playerC := models.Player{ID: uuid.New(), Wins: 6, Losses: 4}  // Seed 3
	playerD := models.Player{ID: uuid.New(), Wins: 4, Losses: 6}  // Seed 4
	mockPlayers := []models.Player{playerC, playerA, playerD, playerB}

	mockLeague := &models.League{
		ID:     leagueID,
		Status: enums.LeagueStatusPostRegularSeason,
		Format: &models.LeagueFormat{
			SeasonType:              enums.LeagueSeasonTypeHybrid,
			PlayoffType:             enums.LeaguePlayoffTypeSingleElim,
			PlayoffSeedingType:      enums.LeaguePlayoffSeedingTypeStandard,
			GroupCount:              1,
			PlayoffParticipantCount: 4,
			PlayoffByesCount:        0,
		},
	}

	// Define Mock Expectations using .On() with the imported mocks
	mockLeagueRepo.On("GetLeagueByID", leagueID).Return(mockLeague, nil)
	mockPlayerRepo.On("GetPlayersByLeagueAndGroupNumber", leagueID, 1).Return(mockPlayers, nil)
	mockGameRepo.On("CreateGames", mock.AnythingOfType("[]*models.Game")).Return(nil)

	gameService := services.NewGameService(mockGameRepo, mockLeagueRepo, mockPlayerRepo)

	// ACT
	err := gameService.GeneratePlayoffBracket(leagueID)

	// ASSERT
	assert.NoError(t, err)

	// Assert that the mock methods were called as expected
	mockLeagueRepo.AssertExpectations(t)
	mockPlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)

	// Capture and inspect the argument passed to CreateGames
	capturedGames := mockGameRepo.Calls[0].Arguments.Get(0).([]*models.Game)
	assert.Len(t, capturedGames, 3, "Should create 2 first-round games and 1 final")

	round1Games := make([]*models.Game, 0)
	for _, game := range capturedGames {
		if game.RoundNumber == 1 {
			round1Games = append(round1Games, game)
		}
	}
	assert.Len(t, round1Games, 2, "Should be 2 games in round 1")

	// Check matchup 1: Seed 1 (playerA) vs Seed 4 (playerD)
	match1Found := (round1Games[0].Player1ID == playerA.ID && round1Games[0].Player2ID == playerD.ID) ||
		(round1Games[0].Player1ID == playerD.ID && round1Games[0].Player2ID == playerA.ID) ||
		(round1Games[1].Player1ID == playerA.ID && round1Games[1].Player2ID == playerD.ID) ||
		(round1Games[1].Player1ID == playerD.ID && round1Games[1].Player2ID == playerA.ID)
	assert.True(t, match1Found, "Expected matchup between Seed 1 (Player A) and Seed 4 (Player D) was not found")

	// Check matchup 2: Seed 2 (playerB) vs Seed 3 (playerC)
	match2Found := (round1Games[0].Player1ID == playerB.ID && round1Games[0].Player2ID == playerC.ID) ||
		(round1Games[0].Player1ID == playerC.ID && round1Games[0].Player2ID == playerB.ID) ||
		(round1Games[1].Player1ID == playerB.ID && round1Games[1].Player2ID == playerC.ID) ||
		(round1Games[1].Player1ID == playerC.ID && round1Games[1].Player2ID == playerB.ID)
	assert.True(t, match2Found, "Expected matchup between Seed 2 (Player B) and Seed 3 (Player C) was not found")
}

func TestGameService_GenerateRegularSeasonGames_Success(t *testing.T) {
	// ARRANGE
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockPlayerRepo := new(mock_repos.MockPlayerRepository)
	mockGameRepo := new(mock_repos.MockGameRepository)

	leagueID := uuid.New()
	mockPlayers := []models.Player{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	mockLeague := &models.League{
		ID:     leagueID,
		Status: enums.LeagueStatusPostDraft,
		Format: &models.LeagueFormat{
			SeasonType: enums.LeagueSeasonTypeHybrid,
			GroupCount: 1,
		},
	}

	mockLeagueRepo.On("GetLeagueByID", leagueID).Return(mockLeague, nil)
	mockPlayerRepo.On("GetPlayersByLeagueAndGroupNumber", leagueID, 1).Return(mockPlayers, nil)
	// We expect 6 games for a 4-player round-robin.
	mockGameRepo.On("CreateGames", mock.MatchedBy(func(games []*models.Game) bool {
		return len(games) == 6
	})).Return(nil)

	gameService := services.NewGameService(mockGameRepo, mockLeagueRepo, mockPlayerRepo)

	// ACT
	err := gameService.GenerateRegularSeasonGames(leagueID)

	// ASSERT
	assert.NoError(t, err)
	mockLeagueRepo.AssertExpectations(t)
	mockPlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}

func TestGameService_GenerateRegularSeasonGames_ErrInvalidState(t *testing.T) {
	// ARRANGE
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockPlayerRepo := new(mock_repos.MockPlayerRepository)
	mockGameRepo := new(mock_repos.MockGameRepository)

	leagueID := uuid.New()

	// League is in DRAFTING status, not POST_DRAFT, which should cause an error.
	mockLeague := &models.League{
		ID:     leagueID,
		Status: enums.LeagueStatusDrafting,
		Format: &models.LeagueFormat{
			SeasonType: enums.LeagueSeasonTypeBracketOnly,
			GroupCount: 1,
		},
	}

	mockLeagueRepo.On("GetLeagueByID", leagueID).Return(mockLeague, nil)
	gameService := services.NewGameService(mockGameRepo, mockLeagueRepo, mockPlayerRepo)

	// ACT
	err := gameService.GenerateRegularSeasonGames(leagueID)

	// ASSERT
	assert.Error(t, err)
	// We can check for a specific error if the service returns a custom error type
	// For now, just asserting that an error is returned is sufficient.
	mockLeagueRepo.AssertExpectations(t)
	// Other repos should not have been called.
	mockPlayerRepo.AssertNotCalled(t, "GetPlayersByLeagueAndGroupNumber")
	mockGameRepo.AssertNotCalled(t, "CreateGames")
}

func TestGameService_GeneratePlayoffBracket_ErrInvalidConfig(t *testing.T) {
	// ARRANGE
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockPlayerRepo := new(mock_repos.MockPlayerRepository)
	mockGameRepo := new(mock_repos.MockGameRepository)

	leagueID := uuid.New()
	mockPlayers := []models.Player{{ID: uuid.New()}, {ID: uuid.New()}}

	// This is an invalid configuration: Single Elimination cannot be Fully Seeded.
	mockLeague := &models.League{
		ID:     leagueID,
		Status: enums.LeagueStatusPostRegularSeason,
		Format: &models.LeagueFormat{
			SeasonType:              enums.LeagueSeasonTypeHybrid,
			PlayoffType:             enums.LeaguePlayoffTypeSingleElim,
			PlayoffSeedingType:      enums.LeaguePlayoffSeedingTypeFullySeeded,
			GroupCount:              1,
			PlayoffParticipantCount: 2,
		},
	}

	mockLeagueRepo.On("GetLeagueByID", leagueID).Return(mockLeague, nil)
	mockPlayerRepo.On("GetPlayersByLeagueAndGroupNumber", leagueID, 1).Return(mockPlayers, nil)

	gameService := services.NewGameService(mockGameRepo, mockLeagueRepo, mockPlayerRepo)

	// ACT
	err := gameService.GeneratePlayoffBracket(leagueID)

	// ASSERT
	assert.Error(t, err)
	mockLeagueRepo.AssertExpectations(t)
	mockPlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertNotCalled(t, "CreateGames")
}
