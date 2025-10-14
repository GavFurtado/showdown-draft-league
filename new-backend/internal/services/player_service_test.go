package services_test

import (
	// "errors"
	"fmt"
	"testing"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupPlayerServiceTest() (services.PlayerService, *mock_repositories.MockPlayerRepository, *mock_repositories.MockLeagueRepository, *mock_repositories.MockUserRepository) {
	mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)

	service := services.NewPlayerService(mockPlayerRepo, mockLeagueRepo, mockUserRepo)
	return service, mockPlayerRepo, mockLeagueRepo, mockUserRepo
}

func TestPlayerService_CreatePlayerHandler(t *testing.T) {
	service, mockPlayerRepo, mockLeagueRepo, mockUserRepo := setupPlayerServiceTest()

	leagueID := uuid.New()
	userID := uuid.New()
	discordUsername := "testuser"
	inLeagueName := "Test Player"
	teamName := "The Testers"

	input := &common.PlayerCreateRequest{
		LeagueID:     leagueID,
		UserID:       userID,
		InLeagueName: &inLeagueName,
		TeamName:     &teamName,
	}

	testLeague := &models.League{ID: leagueID, StartingDraftPoints: 100}
	testUser := &models.User{ID: userID, DiscordUsername: discordUsername}

	t.Run("Success - Create Player", func(t *testing.T) {
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(testLeague, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return(testUser, nil).Once()
		mockPlayerRepo.On("FindPlayerByUserAndLeague", userID, leagueID).Return(nil, nil).Once()
		mockPlayerRepo.On("FindPlayerByInLeagueNameAndLeagueID", *input.InLeagueName, leagueID).Return(nil, nil).Once()
		mockPlayerRepo.On("FindPlayerByTeamNameAndLeagueID", *input.TeamName, leagueID).Return(nil, nil).Once()

		expectedPlayer := &models.Player{
			UserID:       userID,
			LeagueID:     leagueID,
			InLeagueName: *input.InLeagueName,
			TeamName:     *input.TeamName,
			DraftPoints:  int(testLeague.StartingDraftPoints),
			Wins:         0,
			Losses:       0,
			Role:         rbac.PRoleMember,
		}
		createdPlayer := *expectedPlayer
		createdPlayer.ID = uuid.New()
		mockPlayerRepo.On("CreatePlayer", expectedPlayer).Return(&createdPlayer, nil).Once()

		result, err := service.CreatePlayerHandler(input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, createdPlayer.ID, result.ID)
		assert.Equal(t, *input.InLeagueName, result.InLeagueName)

		mockPlayerRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Success - Name Fallback", func(t *testing.T) {
		fallbackInput := &common.PlayerCreateRequest{
			LeagueID:     leagueID,
			UserID:       userID,
			InLeagueName: nil, // Test fallback
			TeamName:     nil, // Test fallback
		}

		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(testLeague, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return(testUser, nil).Once()
		mockPlayerRepo.On("FindPlayerByUserAndLeague", userID, leagueID).Return(nil, nil).Once()
		mockPlayerRepo.On("FindPlayerByInLeagueNameAndLeagueID", discordUsername, leagueID).Return(nil, nil).Once()
		mockPlayerRepo.On("FindPlayerByTeamNameAndLeagueID", discordUsername, leagueID).Return(nil, nil).Once()

		expectedPlayer := &models.Player{
			UserID:       userID,
			LeagueID:     leagueID,
			InLeagueName: discordUsername, // Should fallback
			TeamName:     discordUsername, // Should fallback
			DraftPoints:  int(testLeague.StartingDraftPoints),
			Wins:         0,
			Losses:       0,
			Role:         rbac.PRoleMember,
		}
		createdPlayer := *expectedPlayer
		createdPlayer.ID = uuid.New()
		mockPlayerRepo.On("CreatePlayer", expectedPlayer).Return(&createdPlayer, nil).Once()

		result, err := service.CreatePlayerHandler(fallbackInput)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, discordUsername, result.InLeagueName)
		assert.Equal(t, discordUsername, result.TeamName)

		mockPlayerRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - League Not Found", func(t *testing.T) {
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(nil, gorm.ErrRecordNotFound).Once()

		result, err := service.CreatePlayerHandler(input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, common.ErrLeagueNotFound, err)

		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertNotCalled(t, "GetUserByID")
		mockPlayerRepo.AssertNotCalled(t, "CreatePlayer")
	})

	t.Run("Failure - User Already In League", func(t *testing.T) {
		existingPlayer := &models.Player{ID: uuid.New()}
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(testLeague, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return(testUser, nil).Once()
		mockPlayerRepo.On("FindPlayerByUserAndLeague", userID, leagueID).Return(existingPlayer, nil).Once()

		result, err := service.CreatePlayerHandler(input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, common.ErrUserAlreadyInLeague, err)

		mockPlayerRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - InLeagueName Taken", func(t *testing.T) {
		existingPlayer := &models.Player{ID: uuid.New()}
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(testLeague, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return(testUser, nil).Once()
		mockPlayerRepo.On("FindPlayerByUserAndLeague", userID, leagueID).Return(nil, nil).Once()
		mockPlayerRepo.On("FindPlayerByInLeagueNameAndLeagueID", *input.InLeagueName, leagueID).Return(existingPlayer, nil).Once()

		result, err := service.CreatePlayerHandler(input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, common.ErrInLeagueNameTaken)
		assert.Contains(t, err.Error(), *input.InLeagueName)

		mockPlayerRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - TeamName Taken", func(t *testing.T) {
		existingPlayer := &models.Player{ID: uuid.New()}
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(testLeague, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return(testUser, nil).Once()
		mockPlayerRepo.On("FindPlayerByUserAndLeague", userID, leagueID).Return(nil, nil).Once()
		mockPlayerRepo.On("FindPlayerByInLeagueNameAndLeagueID", *input.InLeagueName, leagueID).Return(nil, nil).Once()
		mockPlayerRepo.On("FindPlayerByTeamNameAndLeagueID", *input.TeamName, leagueID).Return(existingPlayer, nil).Once()

		result, err := service.CreatePlayerHandler(input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, common.ErrTeamNameTaken)
		assert.Contains(t, err.Error(), *input.TeamName)

		mockPlayerRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

//	func TestPlayerService_UpdatePlayerProfile(t *testing.T) {
//		service, mockPlayerRepo, _, _ := setupPlayerServiceTest()
//
//		playerID := uuid.New()
//		userID := uuid.New()
//		leagueID := uuid.New()
//		adminUserID := uuid.New()
//		ownerUserID := uuid.New()
//
//		currentUser := &models.User{ID: userID, Role: "user"}
//		adminUser := &models.User{ID: adminUserID, Role: "admin"}
//		ownerUser := &models.User{ID: ownerUserID, Role: "user"}
//
//		existingPlayer := &models.Player{
//			ID:           playerID,
//			UserID:       userID,
//			LeagueID:     leagueID,
//			InLeagueName: "OldInLeagueName",
//			TeamName:     "OldTeamName",
//			Role:         rbac.PRoleMember,
//		}
//
//		ownerPlayer := &models.Player{
//			ID:       uuid.New(),
//			UserID:   ownerUserID,
//			LeagueID: leagueID,
//			Role:     rbac.PRoleOwner,
//		}
//
//		newInLeagueName := "NewInLeagueName"
//		newTeamName := "NewTeamName"
//
//		t.Run("Success - User updates their own profile", func(t *testing.T) {
//			mockPlayerRepo.On("GetPlayerByID", playerID).Return(existingPlayer, nil).Once()
//			mockPlayerRepo.On("FindPlayerByInLeagueNameAndLeagueID", newInLeagueName, leagueID).Return(nil, nil).Once()
//			mockPlayerRepo.On("FindPlayerByTeamNameAndLeagueID", newTeamName, leagueID).Return(nil, nil).Once()
//			mockPlayerRepo.On("UpdatePlayer", mock.AnythingOfType("*models.Player")).Return(
//				&models.Player{
//					ID:           playerID,
//					UserID:       userID,
//					LeagueID:     leagueID,
//					InLeagueName: newInLeagueName,
//					TeamName:     newTeamName,
//				}, nil).Once()
//
//			result, err := service.UpdatePlayerProfile(currentUser, playerID, &newInLeagueName, &newTeamName)
//
//			assert.NoError(t, err)
//			assert.NotNil(t, result)
//			assert.Equal(t, newInLeagueName, result.InLeagueName)
//			assert.Equal(t, newTeamName, result.TeamName)
//
//			mockPlayerRepo.AssertExpectations(t)
//		})
//
//		t.Run("Success - Admin updates a player's profile", func(t *testing.T) {
//			mockPlayerRepo.On("GetPlayerByID", playerID).Return(existingPlayer, nil).Once()
//			mockPlayerRepo.On("FindPlayerByInLeagueNameAndLeagueID", newInLeagueName, leagueID).Return(nil, nil).Once()
//			mockPlayerRepo.On("UpdatePlayer", mock.AnythingOfType("*models.Player")).Return(&models.Player{InLeagueName: newInLeagueName}, nil).Once()
//
//			_, err := service.UpdatePlayerProfile(adminUser, playerID, &newInLeagueName, nil)
//
//			assert.NoError(t, err)
//			mockPlayerRepo.AssertExpectations(t)
//		})
//
//		t.Run("Success - League Owner updates a player's profile", func(t *testing.T) {
//			mockPlayerRepo.On("GetPlayerByID", playerID).Return(existingPlayer, nil).Once()
//			mockPlayerRepo.On("GetPlayerByUserAndLeague", ownerUser.ID, leagueID).Return(ownerPlayer, nil).Once()
//			mockPlayerRepo.On("FindPlayerByInLeagueNameAndLeagueID", newInLeagueName, leagueID).Return(nil, nil).Once()
//			mockPlayerRepo.On("UpdatePlayer", mock.AnythingOfType("*models.Player")).Return(&models.Player{InLeagueName: newInLeagueName}, nil).Once()
//
//			_, err := service.UpdatePlayerProfile(ownerUser, playerID, &newInLeagueName, nil)
//
//			assert.NoError(t, err)
//			mockPlayerRepo.AssertExpectations(t)
//		})
//
//		t.Run("Failure - Unauthorized user", func(t *testing.T) {
//			unauthorizedUser := &models.User{ID: uuid.New(), Role: "user"}
//			unauthorizedPlayer := &models.Player{ID: uuid.New(), UserID: unauthorizedUser.ID, LeagueID: leagueID, Role: rbac.PRoleMember}
//
//			mockPlayerRepo.On("GetPlayerByID", playerID).Return(existingPlayer, nil).Once()
//			mockPlayerRepo.On("GetPlayerByUserAndLeague", unauthorizedUser.ID, leagueID).Return(unauthorizedPlayer, nil).Once()
//
//			result, err := service.UpdatePlayerProfile(unauthorizedUser, playerID, &newInLeagueName, &newTeamName)
//
//			assert.Error(t, err)
//			assert.Nil(t, result)
//			assert.Equal(t, common.ErrUnauthorized, err)
//			mockPlayerRepo.AssertExpectations(t)
//		})
//
//		t.Run("Failure - Player not found", func(t *testing.T) {
//			mockPlayerRepo.On("GetPlayerByID", playerID).Return(nil, gorm.ErrRecordNotFound).Once()
//
//			result, err := service.UpdatePlayerProfile(currentUser, playerID, &newInLeagueName, &newTeamName)
//
//			assert.Error(t, err)
//			assert.Nil(t, result)
//			assert.Equal(t, common.ErrPlayerNotFound, err)
//			mockPlayerRepo.AssertExpectations(t)
//		})
//
//		t.Run("Failure - InLeagueName is taken", func(t *testing.T) {
//			conflictingPlayer := &models.Player{ID: uuid.New()}
//			mockPlayerRepo.On("GetPlayerByID", playerID).Return(existingPlayer, nil).Once()
//			mockPlayerRepo.On("FindPlayerByInLeagueNameAndLeagueID", newInLeagueName, leagueID).Return(conflictingPlayer, nil).Once()
//
//			result, err := service.UpdatePlayerProfile(currentUser, playerID, &newInLeagueName, nil)
//
//			assert.Error(t, err)
//			assert.Nil(t, result)
//			assert.ErrorIs(t, err, common.ErrInLeagueNameTaken)
//			mockPlayerRepo.AssertExpectations(t)
//		})
//	}
//
//	func TestPlayerService_UpdatePlayerRole(t *testing.T) {
//		service, mockPlayerRepo, _, _ := setupPlayerServiceTest()
//		currentUserID := uuid.New()
//		playerID := uuid.New()
//		newRole := rbac.PRoleModerator
//
//		t.Run("Success - Update Player Role", func(t *testing.T) {
//			updatedPlayer := &models.Player{ID: playerID, Role: newRole}
//			mockPlayerRepo.On("UpdatePlayerRole", playerID, newRole).Return(nil).Once()
//			mockPlayerRepo.On("GetPlayerByID", playerID).Return(updatedPlayer, nil).Once()
//
//			result, err := service.UpdatePlayerRole(currentUserID, playerID, newRole)
//			assert.NoError(t, err)
//			assert.NotNil(t, result)
//			assert.Equal(t, newRole, result.Role)
//			mockPlayerRepo.AssertExpectations(t)
//		})
//
//		t.Run("Failure - Player not found on update", func(t *testing.T) {
//			mockPlayerRepo.On("UpdatePlayerRole", playerID, newRole).Return(gorm.ErrRecordNotFound).Once()
//
//			result, err := service.UpdatePlayerRole(currentUserID, playerID, newRole)
//			assert.Error(t, err)
//			assert.Nil(t, result)
//			assert.Equal(t, common.ErrPlayerNotFound, err)
//			mockPlayerRepo.AssertExpectations(t)
//		})
//
//		t.Run("Failure - DB error on update", func(t *testing.T) {
//			dbError := errors.New("db error")
//			mockPlayerRepo.On("UpdatePlayerRole", playerID, newRole).Return(dbError).Once()
//
//			result, err := service.UpdatePlayerRole(currentUserID, playerID, newRole)
//			assert.Error(t, err)
//			assert.Nil(t, result)
//			assert.ErrorIs(t, err, common.ErrInternalService)
//			mockPlayerRepo.AssertExpectations(t)
//		})
//
//		t.Run("Failure - DB error on re-fetch", func(t *testing.T) {
//			dbError := errors.New("db error")
//			mockPlayerRepo.On("UpdatePlayerRole", playerID, newRole).Return(nil).Once()
//			mockPlayerRepo.On("GetPlayerByID", playerID).Return(nil, dbError).Once()
//
//			result, err := service.UpdatePlayerRole(currentUserID, playerID, newRole)
//			assert.Error(t, err)
//			assert.Nil(t, result)
//			assert.ErrorIs(t, err, common.ErrInternalService)
//			mockPlayerRepo.AssertExpectations(t)
//		})
//	}
//
// // Helper function for testing update methods with similar auth logic
func testPlayerUpdateByAuthorizedUser(t *testing.T, methodName string, updateFunc func(service services.PlayerService, user *models.User, playerID uuid.UUID) (*models.Player, error)) {
	service, mockPlayerRepo, _, _ := setupPlayerServiceTest()

	playerID := uuid.New()
	leagueID := uuid.New()
	adminUserID := uuid.New()
	ownerUserID := uuid.New()
	modUserID := uuid.New()
	unauthorizedUserID := uuid.New()

	adminUser := &models.User{ID: adminUserID, Role: "admin"}
	ownerUser := &models.User{ID: ownerUserID, Role: "user"}
	modUser := &models.User{ID: modUserID, Role: "user"}
	unauthorizedUser := &models.User{ID: unauthorizedUserID, Role: "user"}

	existingPlayer := &models.Player{ID: playerID, UserID: uuid.New(), LeagueID: leagueID, Role: rbac.PRoleMember}
	ownerPlayer := &models.Player{UserID: ownerUserID, LeagueID: leagueID, Role: rbac.PRoleOwner}
	modPlayer := &models.Player{UserID: modUserID, LeagueID: leagueID, Role: rbac.PRoleModerator}
	unauthorizedPlayer := &models.Player{UserID: unauthorizedUserID, LeagueID: leagueID, Role: rbac.PRoleMember}

	// Mock repository calls common to all authorized users
	setupSuccessMocks := func() {
		switch methodName {
		case "UpdatePlayerDraftPoints":
			mockPlayerRepo.On("UpdatePlayerDraftPoints", playerID, 120).Return(nil).Once()
		case "UpdatePlayerRecord":
			mockPlayerRepo.On("UpdatePlayerRecord", playerID, 5, 2).Return(nil).Once()
		case "UpdatePlayerDraftPosition":
			mockPlayerRepo.On("UpdatePlayerDraftPosition", playerID, 1).Return(nil).Once()
		}

		mockPlayerRepo.On("GetPlayerByID", playerID).Return(existingPlayer, nil).Once() // for re-fetch
	}

	t.Run(fmt.Sprintf("Success - %s by Admin", methodName), func(t *testing.T) {
		mockPlayerRepo.On("GetPlayerByID", playerID).Return(existingPlayer, nil).Once() // for auth check
		setupSuccessMocks()
		_, err := updateFunc(service, adminUser, playerID)
		assert.NoError(t, err)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run(fmt.Sprintf("Success - %s by League Owner", methodName), func(t *testing.T) {
		mockPlayerRepo.On("GetPlayerByID", playerID).Return(existingPlayer, nil).Once() // for auth check
		mockPlayerRepo.On("GetPlayerByUserAndLeague", ownerUser.ID, leagueID).Return(ownerPlayer, nil).Once()
		setupSuccessMocks()
		_, err := updateFunc(service, ownerUser, playerID)
		assert.NoError(t, err)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run(fmt.Sprintf("Success - %s by League Moderator", methodName), func(t *testing.T) {
		mockPlayerRepo.On("GetPlayerByID", playerID).Return(existingPlayer, nil).Once() // for auth check
		mockPlayerRepo.On("GetPlayerByUserAndLeague", modUser.ID, leagueID).Return(modPlayer, nil).Once()
		setupSuccessMocks()
		_, err := updateFunc(service, modUser, playerID)
		assert.NoError(t, err)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run(fmt.Sprintf("Failure - %s by Unauthorized User", methodName), func(t *testing.T) {
		mockPlayerRepo.On("GetPlayerByID", playerID).Return(existingPlayer, nil).Once() // for auth check
		mockPlayerRepo.On("GetPlayerByUserAndLeague", unauthorizedUser.ID, leagueID).Return(unauthorizedPlayer, nil).Once()

		_, err := updateFunc(service, unauthorizedUser, playerID)
		assert.Error(t, err)
		assert.Equal(t, common.ErrUnauthorized, err)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestPlayerService_UpdatePlayerDraftPoints(t *testing.T) {
	newPoints := 120
	updateFunc := func(service services.PlayerService, user *models.User, playerID uuid.UUID) (*models.Player, error) {
		return service.UpdatePlayerDraftPoints(user, playerID, &newPoints)
	}
	testPlayerUpdateByAuthorizedUser(t, "UpdatePlayerDraftPoints", updateFunc)
}

func TestPlayerService_UpdatePlayerRecord(t *testing.T) {
	updateFunc := func(service services.PlayerService, user *models.User, playerID uuid.UUID) (*models.Player, error) {
		return service.UpdatePlayerRecord(user, playerID, 5, 2)
	}
	testPlayerUpdateByAuthorizedUser(t, "UpdatePlayerRecord", updateFunc)
}

func TestPlayerService_UpdatePlayerDraftPosition(t *testing.T) {
	updateFunc := func(service services.PlayerService, user *models.User, playerID uuid.UUID) (*models.Player, error) {
		return service.UpdatePlayerDraftPosition(user, playerID, 1)
	}
	testPlayerUpdateByAuthorizedUser(t, "UpdatePlayerDraftPosition", updateFunc)
}

