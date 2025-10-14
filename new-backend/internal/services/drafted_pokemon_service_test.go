package services_test

import (
	"errors"
	"testing"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	mock_repositories "github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestDraftedPokemonService_ReleasePokemon(t *testing.T) {
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)

	service := services.NewDraftedPokemonService(
		mockDraftedPokemonRepo,
		mockUserRepo,
		mockLeagueRepo,
		mockPlayerRepo,
		mockPokemonSpeciesRepo,
		mockLeaguePokemonRepo,
	)

	testDraftedPokemonID := uuid.New()
	testPlayerID := uuid.New()
	testLeagueID := uuid.New()
	testUserID := uuid.New()

	t.Run("Successfully releases a pokemon as admin", func(t *testing.T) {
		currentUser := &models.User{ID: testUserID, Role: "admin"}

		draftedPokemon := &models.DraftedPokemon{
			ID:         testDraftedPokemonID,
			PlayerID:   testPlayerID,
			LeagueID:   testLeagueID,
			IsReleased: false,
		}
		player := &models.Player{
			ID:       testPlayerID,
			UserID:   testUserID,
			LeagueID: testLeagueID,
			Role:     rbac.PRoleMember,
		}

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return(draftedPokemon, nil).Once()
		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return(player, nil).Once()
		// Admin users skip the GetPlayerByUserAndLeague call
		mockDraftedPokemonRepo.On("ReleasePokemon", testDraftedPokemonID).Return(nil).Once()

		err := service.ReleasePokemon(currentUser, testDraftedPokemonID)
		assert.NoError(t, err)

		mockDraftedPokemonRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Successfully releases a pokemon as owner", func(t *testing.T) {
		currentUser := &models.User{ID: testUserID, Role: "user"}

		draftedPokemon := &models.DraftedPokemon{
			ID:         testDraftedPokemonID,
			PlayerID:   testPlayerID,
			LeagueID:   testLeagueID,
			IsReleased: false,
		}
		ownerPlayer := &models.Player{
			ID:       testPlayerID,
			UserID:   testUserID, // Same as current user - they own the pokemon
			LeagueID: testLeagueID,
			Role:     rbac.PRoleMember,
		}

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return(draftedPokemon, nil).Once()
		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return(ownerPlayer, nil).Once()
		mockPlayerRepo.On("GetPlayerByUserAndLeague", currentUser.ID, testLeagueID).Return(ownerPlayer, nil).Once()
		mockDraftedPokemonRepo.On("ReleasePokemon", testDraftedPokemonID).Return(nil).Once()

		err := service.ReleasePokemon(currentUser, testDraftedPokemonID)
		assert.NoError(t, err)

		mockDraftedPokemonRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Fails to release an already released pokemon", func(t *testing.T) {
		currentUser := &models.User{ID: testUserID, Role: "admin"}

		draftedPokemon := &models.DraftedPokemon{
			ID:         testDraftedPokemonID,
			PlayerID:   testPlayerID,
			LeagueID:   testLeagueID,
			IsReleased: true, // Already released
		}
		// No need to mock GetPlayerByID since it returns early after checking IsReleased

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return(draftedPokemon, nil).Once()
		// Early return - no other calls expected

		err := service.ReleasePokemon(currentUser, testDraftedPokemonID)
		assert.Error(t, err)
		assert.Equal(t, common.ErrPokemonAlreadyReleased, err)

		mockDraftedPokemonRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Fails if drafted pokemon not found", func(t *testing.T) {
		currentUser := &models.User{ID: testUserID, Role: "admin"}

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return((*models.DraftedPokemon)(nil), gorm.ErrRecordNotFound).Once()

		err := service.ReleasePokemon(currentUser, testDraftedPokemonID)
		assert.ErrorIs(t, err, common.ErrDraftedPokemonNotFound)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if getting drafted pokemon returns internal error", func(t *testing.T) {
		currentUser := &models.User{ID: testUserID, Role: "admin"}

		internalErr := errors.New("database connection failed")
		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return((*models.DraftedPokemon)(nil), internalErr).Once()

		err := service.ReleasePokemon(currentUser, testDraftedPokemonID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, common.ErrInternalService)
		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if getting owner player returns internal error", func(t *testing.T) {
		currentUser := &models.User{ID: testUserID, Role: "admin"}

		draftedPokemon := &models.DraftedPokemon{
			ID:         testDraftedPokemonID,
			PlayerID:   testPlayerID,
			LeagueID:   testLeagueID,
			IsReleased: false,
		}
		internalErr := errors.New("player db error")

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return(draftedPokemon, nil).Once()
		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return((*models.Player)(nil), internalErr).Once()
		// Early return due to error - no ReleasePokemon call expected

		err := service.ReleasePokemon(currentUser, testDraftedPokemonID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, common.ErrInternalService)
		mockDraftedPokemonRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Fails if current user is not owner and not admin", func(t *testing.T) {
		// Different user trying to release someone else's pokemon
		otherUserID := uuid.New()
		otherUser := &models.User{ID: otherUserID, Role: "user"} // Not admin

		draftedPokemon := &models.DraftedPokemon{
			ID:         testDraftedPokemonID,
			PlayerID:   testPlayerID, // Owned by testPlayerID
			LeagueID:   testLeagueID,
			IsReleased: false,
		}

		ownerPlayer := &models.Player{
			ID:       testPlayerID,
			UserID:   testUserID, // Original owner (different from otherUserID)
			LeagueID: testLeagueID,
			Role:     rbac.PRoleMember,
		}
		requesterPlayer := &models.Player{
			ID:       uuid.New(), // Different player ID
			UserID:   otherUserID,
			LeagueID: testLeagueID,
			Role:     rbac.PRoleMember, // Not owner/moderator
		}

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return(draftedPokemon, nil).Once()
		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return(ownerPlayer, nil).Once()
		mockPlayerRepo.On("GetPlayerByUserAndLeague", otherUserID, testLeagueID).Return(requesterPlayer, nil).Once()
		// No ReleasePokemon call expected since authorization fails

		err := service.ReleasePokemon(otherUser, testDraftedPokemonID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, common.ErrUnauthorized)
		mockDraftedPokemonRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Fails if current user not found in league", func(t *testing.T) {
		currentUser := &models.User{ID: testUserID, Role: "user"} // Not admin

		draftedPokemon := &models.DraftedPokemon{
			ID:         testDraftedPokemonID,
			PlayerID:   testPlayerID,
			LeagueID:   testLeagueID,
			IsReleased: false,
		}
		ownerPlayer := &models.Player{
			ID:       testPlayerID,
			UserID:   uuid.New(), // Different user owns the pokemon
			LeagueID: testLeagueID,
			Role:     rbac.PRoleMember,
		}

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return(draftedPokemon, nil).Once()
		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return(ownerPlayer, nil).Once()
		mockPlayerRepo.On("GetPlayerByUserAndLeague", currentUser.ID, testLeagueID).Return((*models.Player)(nil), gorm.ErrRecordNotFound).Once()

		err := service.ReleasePokemon(currentUser, testDraftedPokemonID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, common.ErrPlayerNotFound)
		mockDraftedPokemonRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Fails if ReleasePokemon repo call returns error", func(t *testing.T) {
		currentUser := &models.User{ID: testUserID, Role: "admin"}

		draftedPokemon := &models.DraftedPokemon{
			ID:         testDraftedPokemonID,
			PlayerID:   testPlayerID,
			LeagueID:   testLeagueID,
			IsReleased: false,
		}
		player := &models.Player{
			ID:       testPlayerID,
			UserID:   testUserID,
			LeagueID: testLeagueID,
			Role:     rbac.PRoleMember,
		}
		repoErr := errors.New("failed to update db")

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return(draftedPokemon, nil).Once()
		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return(player, nil).Once()
		// Admin user, so no GetPlayerByUserAndLeague call
		mockDraftedPokemonRepo.On("ReleasePokemon", testDraftedPokemonID).Return(repoErr).Once()

		err := service.ReleasePokemon(currentUser, testDraftedPokemonID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, common.ErrInternalService)
		mockDraftedPokemonRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Fails if getting current player returns internal error", func(t *testing.T) {
		currentUser := &models.User{ID: uuid.New(), Role: "user"} // Different from owner

		draftedPokemon := &models.DraftedPokemon{
			ID:         testDraftedPokemonID,
			PlayerID:   testPlayerID,
			LeagueID:   testLeagueID,
			IsReleased: false,
		}
		ownerPlayer := &models.Player{
			ID:       testPlayerID,
			UserID:   testUserID, // Different owner
			LeagueID: testLeagueID,
			Role:     rbac.PRoleMember,
		}

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return(draftedPokemon, nil).Once()
		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return(ownerPlayer, nil).Once()
		mockPlayerRepo.On("GetPlayerByUserAndLeague", currentUser.ID, testLeagueID).Return((*models.Player)(nil), errors.New("db error")).Once()

		err := service.ReleasePokemon(currentUser, testDraftedPokemonID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, common.ErrInternalService)
	})

	t.Run("Successfully releases pokemon as league moderator", func(t *testing.T) {
		currentUser := &models.User{ID: uuid.New(), Role: "user"}

		draftedPokemon := &models.DraftedPokemon{
			ID:         testDraftedPokemonID,
			PlayerID:   testPlayerID,
			LeagueID:   testLeagueID,
			IsReleased: false,
		}
		ownerPlayer := &models.Player{
			ID:       testPlayerID,
			UserID:   testUserID, // Different from current user
			LeagueID: testLeagueID,
			Role:     rbac.PRoleMember,
		}
		moderatorPlayer := &models.Player{
			ID:       uuid.New(),
			UserID:   currentUser.ID,
			LeagueID: testLeagueID,
			Role:     rbac.PRoleModerator, // Has elevated permissions
		}

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testDraftedPokemonID).Return(draftedPokemon, nil).Once()
		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return(ownerPlayer, nil).Once()
		mockPlayerRepo.On("GetPlayerByUserAndLeague", currentUser.ID, testLeagueID).Return(moderatorPlayer, nil).Once()
		mockDraftedPokemonRepo.On("ReleasePokemon", testDraftedPokemonID).Return(nil).Once()

		err := service.ReleasePokemon(currentUser, testDraftedPokemonID)
		assert.NoError(t, err)
	})
}

func TestDraftedPokemonService_GetDraftedPokemonByID(t *testing.T) {
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)

	service := services.NewDraftedPokemonService(
		mockDraftedPokemonRepo,
		mockUserRepo,
		mockLeagueRepo,
		mockPlayerRepo,
		mockPokemonSpeciesRepo,
		mockLeaguePokemonRepo,
	)

	testID := uuid.New()

	t.Run("Successfully gets drafted pokemon by ID", func(t *testing.T) {
		expectedPokemon := &models.DraftedPokemon{
			ID:       testID,
			PlayerID: uuid.New(),
			LeagueID: uuid.New(),
		}

		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testID).Return(expectedPokemon, nil).Once()

		result, err := service.GetDraftedPokemonByID(testID)
		assert.NoError(t, err)
		assert.Equal(t, expectedPokemon, result)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if drafted pokemon not found", func(t *testing.T) {
		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testID).Return((*models.DraftedPokemon)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetDraftedPokemonByID(testID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, common.ErrDraftedPokemonNotFound)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if repository returns internal error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		mockDraftedPokemonRepo.On("GetDraftedPokemonByID", testID).Return((*models.DraftedPokemon)(nil), repoErr).Once()

		result, err := service.GetDraftedPokemonByID(testID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, common.ErrInternalService)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})
}

func TestDraftedPokemonService_GetDraftedPokemonByPlayer(t *testing.T) {
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)

	service := services.NewDraftedPokemonService(
		mockDraftedPokemonRepo,
		mockUserRepo,
		mockLeagueRepo,
		mockPlayerRepo,
		mockPokemonSpeciesRepo,
		mockLeaguePokemonRepo,
	)

	testPlayerID := uuid.New()

	t.Run("Successfully gets drafted pokemon by player", func(t *testing.T) {
		targetPlayer := &models.Player{
			ID:       testPlayerID,
			UserID:   uuid.New(),
			LeagueID: uuid.New(),
		}
		expectedPokemon := []models.DraftedPokemon{
			{ID: uuid.New(), PlayerID: testPlayerID},
			{ID: uuid.New(), PlayerID: testPlayerID},
		}

		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return(targetPlayer, nil).Once()
		mockDraftedPokemonRepo.On("GetDraftedPokemonByPlayer", testPlayerID).Return(expectedPokemon, nil).Once()

		result, err := service.GetDraftedPokemonByPlayer(testPlayerID)
		assert.NoError(t, err)
		assert.Equal(t, expectedPokemon, result)

		mockPlayerRepo.AssertExpectations(t)
		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if player not found", func(t *testing.T) {
		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return((*models.Player)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetDraftedPokemonByPlayer(testPlayerID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, common.ErrPlayerNotFound)

		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Fails if getting player returns internal error", func(t *testing.T) {
		repoErr := errors.New("database error")
		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return((*models.Player)(nil), repoErr).Once()

		result, err := service.GetDraftedPokemonByPlayer(testPlayerID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, common.ErrInternalService)

		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Fails if getting drafted pokemon returns error", func(t *testing.T) {
		targetPlayer := &models.Player{
			ID:       testPlayerID,
			UserID:   uuid.New(),
			LeagueID: uuid.New(),
		}
		repoErr := errors.New("database error")

		mockPlayerRepo.On("GetPlayerByID", testPlayerID).Return(targetPlayer, nil).Once()
		mockDraftedPokemonRepo.On("GetDraftedPokemonByPlayer", testPlayerID).Return(([]models.DraftedPokemon)(nil), repoErr).Once()

		result, err := service.GetDraftedPokemonByPlayer(testPlayerID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, common.ErrInternalService)

		mockPlayerRepo.AssertExpectations(t)
		mockDraftedPokemonRepo.AssertExpectations(t)
	})
}

func TestDraftedPokemonService_IsPokemonDrafted(t *testing.T) {
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)

	service := services.NewDraftedPokemonService(
		mockDraftedPokemonRepo,
		mockUserRepo,
		mockLeagueRepo,
		mockPlayerRepo,
		mockPokemonSpeciesRepo,
		mockLeaguePokemonRepo,
	)

	testLeagueID := uuid.New()
	testPokemonSpeciesID := int64(25)

	t.Run("Returns true when pokemon is drafted", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", testPokemonSpeciesID).Return(&models.PokemonSpecies{}, nil).Once()
		mockDraftedPokemonRepo.On("IsPokemonDrafted", testLeagueID, testPokemonSpeciesID).Return(true, nil).Once()

		result, err := service.IsPokemonDrafted(testLeagueID, testPokemonSpeciesID)
		assert.NoError(t, err)
		assert.True(t, result)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Returns false when pokemon is not drafted", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", testPokemonSpeciesID).Return(&models.PokemonSpecies{}, nil).Once()
		mockDraftedPokemonRepo.On("IsPokemonDrafted", testLeagueID, testPokemonSpeciesID).Return(false, nil).Once()

		result, err := service.IsPokemonDrafted(testLeagueID, testPokemonSpeciesID)
		assert.NoError(t, err)
		assert.False(t, result)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if repository returns error", func(t *testing.T) {
		repoErr := errors.New("database error")
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", testPokemonSpeciesID).Return(&models.PokemonSpecies{}, nil).Once()
		mockDraftedPokemonRepo.On("IsPokemonDrafted", testLeagueID, testPokemonSpeciesID).Return(false, repoErr).Once()

		result, err := service.IsPokemonDrafted(testLeagueID, testPokemonSpeciesID)
		assert.False(t, result)
		assert.ErrorIs(t, err, common.ErrInternalService)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if pokemon species not found", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", testPokemonSpeciesID).Return(nil, gorm.ErrRecordNotFound).Once()

		result, err := service.IsPokemonDrafted(testLeagueID, testPokemonSpeciesID)
		assert.False(t, result)
		assert.ErrorIs(t, err, common.ErrPokemonSpeciesNotFound)

		mockPokemonSpeciesRepo.AssertExpectations(t)
	})
}

func TestDraftedPokemonService_GetNextDraftPickNumber(t *testing.T) {
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)

	service := services.NewDraftedPokemonService(
		mockDraftedPokemonRepo,
		mockUserRepo,
		mockLeagueRepo,
		mockPlayerRepo,
		mockPokemonSpeciesRepo,
		mockLeaguePokemonRepo,
	)

	testLeagueID := uuid.New()

	t.Run("Successfully gets next draft pick number", func(t *testing.T) {
		expectedPickNumber := 42
		mockLeagueRepo.On("GetLeagueStatus", testLeagueID).Return(enums.LeagueStatusDrafting, nil).Once()
		mockDraftedPokemonRepo.On("GetNextDraftPickNumber", testLeagueID).Return(expectedPickNumber, nil).Once()

		// actual method call
		result, err := service.GetNextDraftPickNumber(testLeagueID)
		assert.NoError(t, err)
		assert.Equal(t, expectedPickNumber, result)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if repository returns error", func(t *testing.T) {
		repoErr := errors.New("database error")
		mockLeagueRepo.On("GetLeagueStatus", testLeagueID).Return(enums.LeagueStatusDrafting, nil).Once()
		mockDraftedPokemonRepo.On("GetNextDraftPickNumber", testLeagueID).Return(0, repoErr).Once()

		result, err := service.GetNextDraftPickNumber(testLeagueID)
		assert.Equal(t, 0, result)
		assert.ErrorIs(t, err, common.ErrInternalService)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if league is not in drafting state", func(t *testing.T) {
		mockLeagueRepo.On("GetLeagueStatus", testLeagueID).Return(enums.LeagueStatusPlayoffs, nil).Once()

		result, err := service.GetNextDraftPickNumber(testLeagueID)
		assert.Equal(t, 0, result)
		assert.ErrorIs(t, err, common.ErrInvalidState)

		mockLeagueRepo.AssertExpectations(t)
	})
}

func TestDraftedPokemonService_GetDraftedPokemonCountByPlayer(t *testing.T) {
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)

	service := services.NewDraftedPokemonService(
		mockDraftedPokemonRepo,
		mockUserRepo,
		mockLeagueRepo,
		mockPlayerRepo,
		mockPokemonSpeciesRepo,
		mockLeaguePokemonRepo,
	)

	testPlayerID := uuid.New()
	currentUser := &models.User{ID: uuid.New(), Role: "user"}

	t.Run("Successfully gets drafted pokemon count", func(t *testing.T) {
		expectedCount := int64(15)
		mockDraftedPokemonRepo.On("GetDraftedPokemonCountByPlayer", testPlayerID).Return(expectedCount, nil).Once()

		result, err := service.GetDraftedPokemonCountByPlayer(currentUser, testPlayerID)
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, result)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if repository returns error", func(t *testing.T) {
		repoErr := errors.New("database error")
		mockDraftedPokemonRepo.On("GetDraftedPokemonCountByPlayer", testPlayerID).Return(int64(0), repoErr).Once()

		result, err := service.GetDraftedPokemonCountByPlayer(currentUser, testPlayerID)
		assert.Equal(t, int64(0), result)
		assert.ErrorIs(t, err, common.ErrInternalService)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})
}

func TestDraftedPokemonService_DeleteDraftedPokemon(t *testing.T) {
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)

	service := services.NewDraftedPokemonService(
		mockDraftedPokemonRepo,
		mockUserRepo,
		mockLeagueRepo,
		mockPlayerRepo,
		mockPokemonSpeciesRepo,
		mockLeaguePokemonRepo,
	)

	testDraftedPokemonID := uuid.New()
	currentUser := &models.User{ID: uuid.New(), Role: "user"}

	t.Run("Successfully deletes drafted pokemon", func(t *testing.T) {
		mockDraftedPokemonRepo.On("DeleteDraftedPokemon", testDraftedPokemonID).Return(nil).Once()

		err := service.DeleteDraftedPokemon(currentUser, testDraftedPokemonID)
		assert.NoError(t, err)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if drafted pokemon not found", func(t *testing.T) {
		mockDraftedPokemonRepo.On("DeleteDraftedPokemon", testDraftedPokemonID).Return(gorm.ErrRecordNotFound).Once()

		err := service.DeleteDraftedPokemon(currentUser, testDraftedPokemonID)
		assert.ErrorIs(t, err, common.ErrDraftedPokemonNotFound)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if repository returns internal error", func(t *testing.T) {
		repoErr := errors.New("database error")
		mockDraftedPokemonRepo.On("DeleteDraftedPokemon", testDraftedPokemonID).Return(repoErr).Once()

		err := service.DeleteDraftedPokemon(currentUser, testDraftedPokemonID)
		assert.ErrorIs(t, err, common.ErrInternalService)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})
}

// Simple getter methods that just pass through to repository
func TestDraftedPokemonService_SimpleGetterMethods(t *testing.T) {
	mockDraftedPokemonRepo := new(mock_repositories.MockDraftedPokemonRepository)
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)

	service := services.NewDraftedPokemonService(
		mockDraftedPokemonRepo,
		mockUserRepo,
		mockLeagueRepo,
		mockPlayerRepo,
		mockPokemonSpeciesRepo,
		mockLeaguePokemonRepo,
	)

	testLeagueID := uuid.New()
	expectedPokemon := []models.DraftedPokemon{
		{ID: uuid.New(), LeagueID: testLeagueID},
		{ID: uuid.New(), LeagueID: testLeagueID},
	}

	t.Run("GetDraftedPokemonByLeague success", func(t *testing.T) {
		mockDraftedPokemonRepo.On("GetDraftedPokemonByLeague", testLeagueID).Return(expectedPokemon, nil).Once()

		result, err := service.GetDraftedPokemonByLeague(testLeagueID)
		assert.NoError(t, err)
		assert.Equal(t, expectedPokemon, result)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("GetDraftedPokemonByLeague error", func(t *testing.T) {
		repoErr := errors.New("database error")
		mockDraftedPokemonRepo.On("GetDraftedPokemonByLeague", testLeagueID).Return(([]models.DraftedPokemon)(nil), repoErr).Once()

		result, err := service.GetDraftedPokemonByLeague(testLeagueID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, common.ErrInternalService)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("GetActiveDraftedPokemonByLeague success", func(t *testing.T) {
		mockDraftedPokemonRepo.On("GetActiveDraftedPokemonByLeague", testLeagueID).Return(expectedPokemon, nil).Once()

		result, err := service.GetActiveDraftedPokemonByLeague(testLeagueID)
		assert.NoError(t, err)
		assert.Equal(t, expectedPokemon, result)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("GetReleasedPokemonByLeague success", func(t *testing.T) {
		mockDraftedPokemonRepo.On("GetReleasedPokemonByLeague", testLeagueID).Return(expectedPokemon, nil).Once()

		result, err := service.GetReleasedPokemonByLeague(testLeagueID)
		assert.NoError(t, err)
		assert.Equal(t, expectedPokemon, result)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})

	t.Run("GetDraftHistory success", func(t *testing.T) {
		mockDraftedPokemonRepo.On("GetDraftHistory", testLeagueID).Return(expectedPokemon, nil).Once()

		result, err := service.GetDraftHistory(testLeagueID)
		assert.NoError(t, err)
		assert.Equal(t, expectedPokemon, result)

		mockDraftedPokemonRepo.AssertExpectations(t)
	})
}
