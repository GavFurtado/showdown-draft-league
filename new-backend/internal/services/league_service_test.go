package services_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLeagueService_CreateLeague(t *testing.T) {
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockDraftRepo := new(mock_repositories.MockDraftRepository)
	mockGameRepo := new(mock_repositories.MockGameRepository)

	service := services.NewLeagueService(
		mockLeagueRepo,
		mockPlayerRepo,
		mockLeaguePokemonRepo,
		mockDraftedPokemonRepo,
		mockDraftRepo,
		mockGameRepo,
	)

	testUserID := uuid.New()
	startDate := time.Now()

	input := &common.LeagueCreateRequestDTO{
		Name:                "Test League",
		RulesetDescription:  "Test rules",
		MaxPokemonPerPlayer: 6,
		StartingDraftPoints: 1000,
		StartDate:           startDate,
		Format: models.LeagueFormat{
			SeasonType:               "ROUND_ROBIN_ONLY",
			GroupCount:               4,
			GamesPerOpponent:         2,
			PlayoffType:              "single",
			PlayoffParticipantCount:  4,
			PlayoffByesCount:         0,
			PlayoffSeedingType:       "regular_season",
			IsSnakeRoundDraft:        true,
			AllowTrading:             true,
			AllowTransferCredits:     false,
			TransferCreditsPerWindow: 0,
		},
	}

	t.Run("Successfully creates league and owner player", func(t *testing.T) {
		expectedLeague := &models.League{
			Name:                input.Name,
			RulesetDescription:  input.RulesetDescription,
			MaxPokemonPerPlayer: input.MaxPokemonPerPlayer,
			StartingDraftPoints: input.StartingDraftPoints,
			StartDate:           input.StartDate,
			Format:              &input.Format,
		}
		createdLeague := *expectedLeague
		createdLeague.ID = uuid.New()

		expectedOwnerPlayer := &models.Player{
			UserID:          testUserID,
			LeagueID:        createdLeague.ID,
			InLeagueName:    "League Owner",
			TeamName:        fmt.Sprintf("%s's Team", input.Name),
			IsParticipating: false,
			DraftPoints:     1000,
			Role:            rbac.PRoleOwner,
		}
		createdPlayer := *expectedOwnerPlayer
		createdPlayer.ID = uuid.New()

		mockLeagueRepo.On("GetLeaguesCountWhereOwner", testUserID).Return(int64(0), nil).Once()
		mockLeagueRepo.On("CreateLeague", expectedLeague).Return(&createdLeague, nil).Once()
		mockPlayerRepo.On("CreatePlayer", expectedOwnerPlayer).Return(&createdPlayer, nil).Once()

		result, err := service.CreateLeague(testUserID, input)
		assert.NoError(t, err)
		assert.Equal(t, &createdLeague, result)
		assert.Equal(t, createdLeague.ID, result.ID)
		assert.Equal(t, input.Name, result.Name)

		mockLeagueRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Fails if user already has maximum leagues", func(t *testing.T) {
		mockLeagueRepo.On("GetLeaguesCountWhereOwner", testUserID).Return(int64(2), nil).Once()

		result, err := service.CreateLeague(testUserID, input)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "max league creation limit reached")

		mockLeagueRepo.AssertExpectations(t)
		mockLeagueRepo.AssertNotCalled(t, "CreateLeague")
		mockPlayerRepo.AssertNotCalled(t, "CreatePlayer")
	})

	t.Run("Fails if GetLeaguesCountWhereOwner returns error", func(t *testing.T) {
		dbError := errors.New("database error")
		mockLeagueRepo.On("GetLeaguesCountWhereOwner", testUserID).Return(int64(0), dbError).Once()

		result, err := service.CreateLeague(testUserID, input)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to check commissioner league count")

		mockLeagueRepo.AssertExpectations(t)
		mockLeagueRepo.AssertNotCalled(t, "CreateLeague")
		mockPlayerRepo.AssertNotCalled(t, "CreatePlayer")
	})

	t.Run("Fails if CreateLeague returns error", func(t *testing.T) {
		dbError := errors.New("database error")
		mockLeagueRepo.On("GetLeaguesCountWhereOwner", testUserID).Return(int64(0), nil).Once()
		mockLeagueRepo.On("CreateLeague", mock.AnythingOfType("*models.League")).Return((*models.League)(nil), dbError).Once()

		result, err := service.CreateLeague(testUserID, input)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create league")

		mockLeagueRepo.AssertExpectations(t)
		mockPlayerRepo.AssertNotCalled(t, "CreatePlayer")
	})

	t.Run("Fails if CreatePlayer returns error", func(t *testing.T) {
		expectedLeague := &models.League{
			Name:                input.Name,
			RulesetDescription:  input.RulesetDescription,
			MaxPokemonPerPlayer: input.MaxPokemonPerPlayer,
			StartingDraftPoints: input.StartingDraftPoints,
			StartDate:           input.StartDate,
			Format:              &input.Format,
		}
		createdLeague := *expectedLeague
		createdLeague.ID = uuid.New()

		dbError := errors.New("player creation error")
		mockLeagueRepo.On("GetLeaguesCountWhereOwner", testUserID).Return(int64(0), nil).Once()
		mockLeagueRepo.On("CreateLeague", expectedLeague).Return(&createdLeague, nil).Once()
		mockPlayerRepo.On("CreatePlayer", mock.AnythingOfType("*models.Player")).Return((*models.Player)(nil), dbError).Once()

		result, err := service.CreateLeague(testUserID, input)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create league owner player")

		mockLeagueRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestLeagueService_GetLeagueByIDForUser(t *testing.T) {
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockDraftRepo := new(mock_repositories.MockDraftRepository)
	mockGameRepo := new(mock_repositories.MockGameRepository)

	service := services.NewLeagueService(
		mockLeagueRepo,
		mockPlayerRepo,
		mockLeaguePokemonRepo,
		mockDraftedPokemonRepo,
		mockDraftRepo,
		mockGameRepo,
	)

	testUserID := uuid.New()
	testLeagueID := uuid.New()

	t.Run("Successfully retrieves league by ID", func(t *testing.T) {
		expectedLeague := &models.League{
			ID:                  testLeagueID,
			Name:                "Test League",
			RulesetDescription:  "Test rules",
			MaxPokemonPerPlayer: 6,
			StartingDraftPoints: 1000,
			StartDate:           time.Now(),
		}

		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(expectedLeague, nil).Once()

		result, err := service.GetLeagueByIDForUser(testUserID, testLeagueID)
		assert.NoError(t, err)
		assert.Equal(t, expectedLeague, result)
		assert.Equal(t, testLeagueID, result.ID)

		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Fails if GetLeagueByID returns error", func(t *testing.T) {
		dbError := errors.New("database error")
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return((*models.League)(nil), dbError).Once()

		result, err := service.GetLeagueByIDForUser(testUserID, testLeagueID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to retrieve league")

		mockLeagueRepo.AssertExpectations(t)
	})
}

func TestLeagueService_GetLeaguesByCommissioner(t *testing.T) {
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockDraftRepo := new(mock_repositories.MockDraftRepository)
	mockGameRepo := new(mock_repositories.MockGameRepository)

	service := services.NewLeagueService(
		mockLeagueRepo,
		mockPlayerRepo,
		mockLeaguePokemonRepo,
		mockDraftedPokemonRepo,
		mockDraftRepo,
		mockGameRepo,
	)

	testUserID := uuid.New()
	currentUser := &models.User{ID: testUserID}

	t.Run("Successfully retrieves commissioner leagues", func(t *testing.T) {
		expectedLeagues := []models.League{
			{
				ID:                  uuid.New(),
				Name:                "League 1",
				RulesetDescription:  "Rules 1",
				MaxPokemonPerPlayer: 6,
				StartingDraftPoints: 1000,
			},
			{
				ID:                  uuid.New(),
				Name:                "League 2",
				RulesetDescription:  "Rules 2",
				MaxPokemonPerPlayer: 8,
				StartingDraftPoints: 1200,
			},
		}

		mockLeagueRepo.On("GetLeaguesByOwner", testUserID).Return(expectedLeagues, nil).Once()

		result, err := service.GetLeaguesByCommissioner(testUserID, currentUser)
		assert.NoError(t, err)
		assert.Equal(t, expectedLeagues, result)
		assert.Len(t, result, 2)

		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Successfully returns empty slice when user owns no leagues", func(t *testing.T) {
		expectedLeagues := []models.League{}

		mockLeagueRepo.On("GetLeaguesByOwner", testUserID).Return(expectedLeagues, nil).Once()

		result, err := service.GetLeaguesByCommissioner(testUserID, currentUser)
		assert.NoError(t, err)
		assert.Equal(t, expectedLeagues, result)
		assert.Empty(t, result)

		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Fails if GetLeaguesByOwner returns error", func(t *testing.T) {
		dbError := errors.New("database error")
		mockLeagueRepo.On("GetLeaguesByOwner", testUserID).Return([]models.League(nil), dbError).Once()

		result, err := service.GetLeaguesByCommissioner(testUserID, currentUser)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to retrieve commissioner leagues")

		mockLeagueRepo.AssertExpectations(t)
	})
}

func TestLeagueService_GetLeaguesByUser(t *testing.T) {
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockDraftRepo := new(mock_repositories.MockDraftRepository)
	mockGameRepo := new(mock_repositories.MockGameRepository)

	service := services.NewLeagueService(
		mockLeagueRepo,
		mockPlayerRepo,
		mockLeaguePokemonRepo,
		mockDraftedPokemonRepo,
		mockDraftRepo,
		mockGameRepo,
	)

	testUserID := uuid.New()
	currentUser := &models.User{ID: testUserID}

	t.Run("Successfully retrieves user leagues", func(t *testing.T) {
		expectedLeagues := []models.League{
			{
				ID:                  uuid.New(),
				Name:                "League A",
				RulesetDescription:  "Rules A",
				MaxPokemonPerPlayer: 6,
				StartingDraftPoints: 1000,
			},
			{
				ID:                  uuid.New(),
				Name:                "League B",
				RulesetDescription:  "Rules B",
				MaxPokemonPerPlayer: 8,
				StartingDraftPoints: 1200,
			},
			{
				ID:                  uuid.New(),
				Name:                "League C",
				RulesetDescription:  "Rules C",
				MaxPokemonPerPlayer: 4,
				StartingDraftPoints: 800,
			},
		}

		mockLeagueRepo.On("GetLeaguesByUser", testUserID).Return(expectedLeagues, nil).Once()

		result, err := service.GetLeaguesByUser(testUserID, currentUser)
		assert.NoError(t, err)
		assert.Equal(t, expectedLeagues, result)
		assert.Len(t, result, 3)

		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Successfully returns empty slice when user is not in any leagues", func(t *testing.T) {
		expectedLeagues := []models.League{}

		mockLeagueRepo.On("GetLeaguesByUser", testUserID).Return(expectedLeagues, nil).Once()

		result, err := service.GetLeaguesByUser(testUserID, currentUser)
		assert.NoError(t, err)
		assert.Equal(t, expectedLeagues, result)
		assert.Empty(t, result)

		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Fails if GetLeaguesByUser returns error", func(t *testing.T) {
		dbError := errors.New("database error")
		mockLeagueRepo.On("GetLeaguesByUser", testUserID).Return([]models.League(nil), dbError).Once()

		result, err := service.GetLeaguesByUser(testUserID, currentUser)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to retrieve leagues")

		mockLeagueRepo.AssertExpectations(t)
	})
}
