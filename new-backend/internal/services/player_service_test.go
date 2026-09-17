package services_test

import (
	"errors"
	"log/slog"
	"testing"

	mock_repos "github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupPlayerServiceTest() (services.PlayerService, *mock_repos.MockPlayerRepository, *mock_repos.MockLeagueRepository, *mock_repos.MockUserRepository) {
	mockPlayerRepo := new(mock_repos.MockPlayerRepository)
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockUserRepo := new(mock_repos.MockUserRepository)

	service := services.NewPlayerService(
		slog.Default(),
		mockPlayerRepo,
		mockLeagueRepo,
		mockUserRepo,
	)

	return service, mockPlayerRepo, mockLeagueRepo, mockUserRepo
}

func TestPlayerService_GetByID(t *testing.T) {
	service, mockPlayerRepo, _, _ := setupPlayerServiceTest()

	t.Run("Success", func(t *testing.T) {
		expected := &models.Player{ID: uuid.New()}
		mockPlayerRepo.On("GetByID", expected.ID).Return(expected, nil).Once()

		result, err := service.GetByID(expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := uuid.New()
		mockPlayerRepo.On("GetByID", id).Return((*models.Player)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetByID(id)
		assert.Error(t, err)
		assert.Equal(t, types.ErrPlayerNotFound, err)
		assert.Nil(t, result)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		id := uuid.New()
		mockPlayerRepo.On("GetByID", id).Return((*models.Player)(nil), errors.New("db error")).Once()

		result, err := service.GetByID(id)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, types.ErrInternalService))
		assert.Nil(t, result)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestPlayerService_GetByUserAndLeague(t *testing.T) {
	service, mockPlayerRepo, _, _ := setupPlayerServiceTest()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		leagueID := uuid.New()
		expected := &models.Player{ID: uuid.New(), UserID: userID, LeagueID: leagueID}
		mockPlayerRepo.On("GetByUserAndLeague", userID, leagueID).Return(expected, nil).Once()

		result, err := service.GetByUserAndLeague(userID, leagueID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockPlayerRepo.On("GetByUserAndLeague", mock.Anything, mock.Anything).Return((*models.Player)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetByUserAndLeague(uuid.New(), uuid.New())
		assert.Error(t, err)
		assert.Equal(t, types.ErrPlayerNotFound, err)
		assert.Nil(t, result)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestPlayerService_GetByLeague(t *testing.T) {
	service, mockPlayerRepo, _, _ := setupPlayerServiceTest()

	t.Run("Success", func(t *testing.T) {
		leagueID := uuid.New()
		expected := []models.Player{{ID: uuid.New(), LeagueID: leagueID}}
		mockPlayerRepo.On("GetByLeague", leagueID).Return(expected, nil).Once()

		result, err := service.GetByLeague(leagueID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockPlayerRepo.On("GetByLeague", mock.Anything).Return([]models.Player(nil), errors.New("db error")).Once()

		result, err := service.GetByLeague(uuid.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, types.ErrInternalService))
		assert.Nil(t, result)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestPlayerService_GetByUser(t *testing.T) {
	service, mockPlayerRepo, _, _ := setupPlayerServiceTest()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		expected := []models.Player{{ID: uuid.New(), UserID: userID}}
		mockPlayerRepo.On("GetByUser", userID).Return(expected, nil).Once()

		result, err := service.GetByUser(userID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockPlayerRepo.On("GetByUser", mock.Anything).Return([]models.Player(nil), errors.New("db error")).Once()

		result, err := service.GetByUser(uuid.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, types.ErrInternalService))
		assert.Nil(t, result)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestPlayerService_GetWithFullRoster(t *testing.T) {
	service, mockPlayerRepo, _, _ := setupPlayerServiceTest()

	t.Run("Success", func(t *testing.T) {
		expected := &models.Player{ID: uuid.New()}
		mockPlayerRepo.On("GetWithFullRoster", expected.ID).Return(expected, nil).Once()

		result, err := service.GetWithFullRoster(expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := uuid.New()
		mockPlayerRepo.On("GetWithFullRoster", id).Return((*models.Player)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetWithFullRoster(id)
		assert.Error(t, err)
		assert.Equal(t, types.ErrPlayerNotFound, err)
		assert.Nil(t, result)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestPlayerService_UpdateProfile(t *testing.T) {
	service, mockPlayerRepo, _, _ := setupPlayerServiceTest()

	t.Run("Success", func(t *testing.T) {
		playerID := uuid.New()
		memberUserID := uuid.New()
		leagueID := uuid.New()
		existing := &models.Player{ID: playerID, UserID: memberUserID, LeagueID: leagueID}

		mockPlayerRepo.On("GetByID", playerID).Return(existing, nil).Once()

		result, err := service.UpdateProfile(&models.User{ID: memberUserID}, playerID, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, existing, result)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("MemberNotFound", func(t *testing.T) {
		mockPlayerRepo.On("GetByID", mock.Anything).Return((*models.Player)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpdateProfile(&models.User{ID: uuid.New()}, uuid.New(), nil, nil)
		assert.Error(t, err)
		assert.Equal(t, types.ErrPlayerNotFound, err)
		assert.Nil(t, result)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestPlayerService_UpdateDraftPoints(t *testing.T) {
	service, mockPlayerRepo, _, _ := setupPlayerServiceTest()

	t.Run("Success", func(t *testing.T) {
		playerID := uuid.New()
		points := 200
		expected := &models.Player{ID: playerID, DraftPoints: points}

		mockPlayerRepo.On("GetByID", playerID).Return(&models.Player{ID: playerID}, nil).Once()
		mockPlayerRepo.On("GetByID", playerID).Return(expected, nil).Once()
		mockPlayerRepo.On("UpdateDraftPoints", playerID, points).Return(nil).Once()

		result, err := service.UpdateDraftPoints(&models.User{ID: uuid.New(), Role: "admin"}, playerID, &points)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestPlayerService_UpdateRecord(t *testing.T) {
	service, mockPlayerRepo, _, _ := setupPlayerServiceTest()

	t.Run("Success", func(t *testing.T) {
		playerID := uuid.New()
		expected := &models.Player{ID: playerID, Wins: 5, Losses: 3}

		mockPlayerRepo.On("GetByID", playerID).Return(&models.Player{ID: playerID}, nil).Once()
		mockPlayerRepo.On("UpdateRecord", playerID, 5, 3).Return(nil).Once()
		mockPlayerRepo.On("GetByID", playerID).Return(expected, nil).Once()

		result, err := service.UpdateRecord(&models.User{ID: uuid.New(), Role: "admin"}, playerID, 5, 3)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestPlayerService_UpdateDraftPosition(t *testing.T) {
	service, mockPlayerRepo, _, _ := setupPlayerServiceTest()

	t.Run("Success", func(t *testing.T) {
		playerID := uuid.New()
		expected := &models.Player{ID: playerID, DraftPosition: 3}

		mockPlayerRepo.On("GetByID", playerID).Return(&models.Player{ID: playerID}, nil).Once()
		mockPlayerRepo.On("UpdateDraftPosition", playerID, 3).Return(nil).Once()
		mockPlayerRepo.On("GetByID", playerID).Return(expected, nil).Once()

		result, err := service.UpdateDraftPosition(&models.User{ID: uuid.New(), Role: "admin"}, playerID, 3)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("MemberNotFound", func(t *testing.T) {
		mockPlayerRepo.On("GetByID", mock.Anything).Return((*models.Player)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpdateDraftPosition(&models.User{ID: uuid.New()}, uuid.New(), 3)
		assert.Error(t, err)
		assert.Equal(t, types.ErrPlayerNotFound, err)
		assert.Nil(t, result)
		mockPlayerRepo.AssertExpectations(t)
	})
}

func TestPlayerService_Create_GroupCountZeroNoPanic(t *testing.T) {
	service, mockPlayerRepo, mockLeagueRepo, mockUserRepo := setupPlayerServiceTest()

	leagueID := uuid.New()
	userID := uuid.New()
	discordUsername := "testuser"
	league := &models.League{
		ID:                  leagueID,
		Status:              enums.LeagueStatusSetup,
		MaxPokemonPerPlayer: 6,
		MinPokemonPerPlayer: 3,
		StartingDraftPoints: 140,
		NewPlayerGroupNumber: 0,
		Format:              &types.LeagueFormat{GroupCount: 0},
	}
	user := &models.User{ID: userID, DiscordUsername: discordUsername}
	player := &models.Player{ID: uuid.New()}

	mockLeagueRepo.On("GetLeagueByID", leagueID).Return(league, nil).Once()
	mockUserRepo.On("GetUserByID", userID).Return(user, nil).Once()
	mockPlayerRepo.On("FindByUserAndLeague", userID, leagueID).Return((*models.Player)(nil), nil).Once()
	mockPlayerRepo.On("FindByInLeagueName", discordUsername, leagueID).Return((*models.Player)(nil), nil).Once()
	mockPlayerRepo.On("FindByTeamName", discordUsername+"'s Team", leagueID).Return((*models.Player)(nil), nil).Once()
	mockPlayerRepo.On("Create", mock.Anything).Return(player, nil).Once()
	mockLeagueRepo.On("UpdateLeague", mock.Anything).Return(league, nil).Once()

	input := &requests.PlayerCreateRequestDTO{
		UserID:   userID,
		LeagueID: leagueID,
	}

	// Must not panic with a GroupCount of 0.
	result, err := service.Create(user, input)
	assert.NoError(t, err)
	assert.Equal(t, player, result)
	mockPlayerRepo.AssertExpectations(t)
	mockLeagueRepo.AssertExpectations(t)
}
