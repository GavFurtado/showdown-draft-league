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
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
)

func TestGameService_GeneratePlayoffBracket_Correct(t *testing.T) {
	// ARRANGE
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockLeagueMemberRepo := new(mock_repos.MockLeagueMemberRepository)
	mockGameRepo := new(mock_repos.MockGameRepository)

	leagueID := uuid.New()

	memberA := models.LeagueMember{ID: uuid.New(), Wins: 10, Losses: 0} // Seed 1
	memberB := models.LeagueMember{ID: uuid.New(), Wins: 8, Losses: 2}  // Seed 2
	memberC := models.LeagueMember{ID: uuid.New(), Wins: 6, Losses: 4}  // Seed 3
	memberD := models.LeagueMember{ID: uuid.New(), Wins: 4, Losses: 6}  // Seed 4
	mockMembers := []models.LeagueMember{memberC, memberA, memberD, memberB}

	mockLeague := &models.League{
		ID:     leagueID,
		Status: enums.LeagueStatusPostRegularSeason,
		Format: &types.LeagueFormat{
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
	mockLeagueMemberRepo.On("GetByLeagueAndGroup", leagueID, 1).Return(mockMembers, nil)
	mockGameRepo.On("CreateGames", mock.AnythingOfType("[]*models.Game")).Return(nil)

	gameService := services.NewGameService(mockGameRepo, mockLeagueRepo, mockLeagueMemberRepo)

	// ACT
	err := gameService.GeneratePlayoffBracket(leagueID)

	// ASSERT
	assert.NoError(t, err)

	// Assert that the mock methods were called as expected
	mockLeagueRepo.AssertExpectations(t)
	mockLeagueMemberRepo.AssertExpectations(t)
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

	// Check matchup 1: Seed 1 (memberA) vs Seed 4 (memberD)
	match1Found := (round1Games[0].Player1ID == memberA.ID && round1Games[0].Player2ID == memberD.ID) ||
		(round1Games[0].Player1ID == memberD.ID && round1Games[0].Player2ID == memberA.ID) ||
		(round1Games[1].Player1ID == memberA.ID && round1Games[1].Player2ID == memberD.ID) ||
		(round1Games[1].Player1ID == memberD.ID && round1Games[1].Player2ID == memberA.ID)
	assert.True(t, match1Found, "Expected matchup between Seed 1 (Member A) and Seed 4 (Member D) was not found")

	// Check matchup 2: Seed 2 (memberB) vs Seed 3 (memberC)
	match2Found := (round1Games[0].Player1ID == memberB.ID && round1Games[0].Player2ID == memberC.ID) ||
		(round1Games[0].Player1ID == memberC.ID && round1Games[0].Player2ID == memberB.ID) ||
		(round1Games[1].Player1ID == memberB.ID && round1Games[1].Player2ID == memberC.ID) ||
		(round1Games[1].Player1ID == memberC.ID && round1Games[1].Player2ID == memberB.ID)
	assert.True(t, match2Found, "Expected matchup between Seed 2 (Member B) and Seed 3 (Member C) was not found")
}

func TestGameService_GenerateRegularSeasonGames_Success(t *testing.T) {
	// ARRANGE
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockLeagueMemberRepo := new(mock_repos.MockLeagueMemberRepository)
	mockGameRepo := new(mock_repos.MockGameRepository)

	leagueID := uuid.New()
	mockMembers := []models.LeagueMember{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	mockLeague := &models.League{
		ID:     leagueID,
		Status: enums.LeagueStatusPostDraft,
		Format: &types.LeagueFormat{
			SeasonType: enums.LeagueSeasonTypeHybrid,
			GroupCount: 1,
		},
	}

	mockLeagueRepo.On("GetLeagueByID", leagueID).Return(mockLeague, nil)
	mockGameRepo.On("HasGames", leagueID, enums.GameTypeRegularSeason).Return(false, nil).Once()
	mockLeagueMemberRepo.On("GetByLeagueAndGroup", leagueID, 1).Return(mockMembers, nil)
	// We expect 6 games for a 4-player round-robin.
	mockGameRepo.On("CreateGames", mock.MatchedBy(func(games []*models.Game) bool {
		return len(games) == 6
	})).Return(nil)

	gameService := services.NewGameService(mockGameRepo, mockLeagueRepo, mockLeagueMemberRepo)

	// ACT
	err := gameService.GenerateRegularSeasonGames(leagueID)

	// ASSERT
	assert.NoError(t, err)
	mockLeagueRepo.AssertExpectations(t)
	mockLeagueMemberRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}

func TestGameService_GenerateRegularSeasonGames_ErrGamesAlreadyGenerated(t *testing.T) {
	// ARRANGE
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockLeagueMemberRepo := new(mock_repos.MockLeagueMemberRepository)
	mockGameRepo := new(mock_repos.MockGameRepository)

	leagueID := uuid.New()
	mockLeague := &models.League{
		ID:     leagueID,
		Status: enums.LeagueStatusPostDraft,
		Format: &types.LeagueFormat{
			SeasonType: enums.LeagueSeasonTypeHybrid,
			GroupCount: 1,
		},
	}

	mockLeagueRepo.On("GetLeagueByID", leagueID).Return(mockLeague, nil)
	mockGameRepo.On("HasGames", leagueID, enums.GameTypeRegularSeason).Return(true, nil).Once()

	gameService := services.NewGameService(mockGameRepo, mockLeagueRepo, mockLeagueMemberRepo)

	// ACT
	err := gameService.GenerateRegularSeasonGames(leagueID)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, types.ErrGamesAlreadyGenerated, err)
	mockLeagueRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockGameRepo.AssertNotCalled(t, "CreateGames")
}

func TestGameService_GenerateRegularSeasonGames_ErrInvalidState(t *testing.T) {
	// ARRANGE
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockLeagueMemberRepo := new(mock_repos.MockLeagueMemberRepository)
	mockGameRepo := new(mock_repos.MockGameRepository)

	leagueID := uuid.New()

	// League is in DRAFTING status, not POST_DRAFT, which should cause an error.
	mockLeague := &models.League{
		ID:     leagueID,
		Status: enums.LeagueStatusDrafting,
		Format: &types.LeagueFormat{
			SeasonType: enums.LeagueSeasonTypeBracketOnly,
			GroupCount: 1,
		},
	}

	mockLeagueRepo.On("GetLeagueByID", leagueID).Return(mockLeague, nil)
	mockGameRepo.On("HasGames", leagueID, enums.GameTypeRegularSeason).Return(false, nil).Once() // Expect this check first

	gameService := services.NewGameService(mockGameRepo, mockLeagueRepo, mockLeagueMemberRepo)

	// ACT
	err := gameService.GenerateRegularSeasonGames(leagueID)

	// ASSERT
	assert.Error(t, err)
	mockLeagueRepo.AssertExpectations(t)
	mockGameRepo.AssertNotCalled(t, "CreateGames")
}

func TestGameService_GeneratePlayoffBracket_ErrInvalidConfig(t *testing.T) {
	// ARRANGE
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockLeagueMemberRepo := new(mock_repos.MockLeagueMemberRepository)
	mockGameRepo := new(mock_repos.MockGameRepository)

	leagueID := uuid.New()
	mockMembers := []models.LeagueMember{{ID: uuid.New()}, {ID: uuid.New()}}

	// This is an invalid configuration: Single Elimination cannot be Fully Seeded.
	mockLeague := &models.League{
		ID:     leagueID,
		Status: enums.LeagueStatusPostRegularSeason,
		Format: &types.LeagueFormat{
			SeasonType:              enums.LeagueSeasonTypeHybrid,
			PlayoffType:             enums.LeaguePlayoffTypeSingleElim,
			PlayoffSeedingType:      enums.LeaguePlayoffSeedingTypeFullySeeded,
			GroupCount:              1,
			PlayoffParticipantCount: 2,
		},
	}

	mockLeagueRepo.On("GetLeagueByID", leagueID).Return(mockLeague, nil)
	mockLeagueMemberRepo.On("GetByLeagueAndGroup", leagueID, 1).Return(mockMembers, nil)

	gameService := services.NewGameService(mockGameRepo, mockLeagueRepo, mockLeagueMemberRepo)

	// ACT
	err := gameService.GeneratePlayoffBracket(leagueID)

	// ASSERT
	assert.Error(t, err)
	mockLeagueRepo.AssertExpectations(t)
	mockLeagueMemberRepo.AssertExpectations(t)
	mockGameRepo.AssertNotCalled(t, "CreateGames")
}
