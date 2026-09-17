package services_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newDevAuthService(mockUserRepo *mock_repositories.MockUserRepository, mockLeagueRepo *mock_repositories.MockLeagueRepository, mockPlayerRepo *mock_repositories.MockPlayerRepository, mockSpectatorRepo *mock_repositories.MockSpectatorRepository) services.DevAuthService {
	jwtService := services.NewJWTService("test-secret")
	return services.NewDevAuthService(slog.Default(), mockUserRepo, mockLeagueRepo, mockPlayerRepo, mockSpectatorRepo, jwtService)
}

func TestDevAuthService_ListUsers(t *testing.T) {
	mockUserRepo := new(mock_repositories.MockUserRepository)
	service := newDevAuthService(mockUserRepo, nil, nil, nil)

	t.Run("Successfully lists users", func(t *testing.T) {
		expectedUsers := []models.User{{DiscordUsername: "alice"}, {DiscordUsername: "bob"}}
		mockUserRepo.On("GetAllUsers").Return(expectedUsers, nil).Once()

		users, err := service.ListUsers()
		assert.NoError(t, err)
		assert.Equal(t, expectedUsers, users)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService for repository errors", func(t *testing.T) {
		mockUserRepo.On("GetAllUsers").Return([]models.User(nil), errors.New("db error")).Once()

		users, err := service.ListUsers()
		assert.ErrorIs(t, err, types.ErrInternalService)
		assert.Nil(t, users)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestDevAuthService_CreateDevUser(t *testing.T) {
	mockUserRepo := new(mock_repositories.MockUserRepository)
	service := newDevAuthService(mockUserRepo, nil, nil, nil)

	t.Run("Successfully creates a dev user", func(t *testing.T) {
		showdown := "testshowdown"
		var created *models.User
		mockUserRepo.On("CreateUser", mock.AnythingOfType("*models.User")).
			Return(func(u *models.User) *models.User {
				u.ID = uuid.New()
				created = u
				return u
			}, nil).Once()

		user, err := service.CreateDevUser("Test User", &showdown)
		assert.NoError(t, err)
		assert.Same(t, created, user)
		assert.Equal(t, "Test User", user.DiscordUsername)
		assert.NotEmpty(t, user.ID)
		assert.Regexp(t, `^dev-[0-9a-f-]{36}$`, user.DiscordID)
		require.NotNil(t, user.ShowdownUsername)
		assert.Equal(t, "testshowdown", *user.ShowdownUsername)
		assert.Equal(t, "user", user.Role)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Creates a user without showdown username for onboarding flow", func(t *testing.T) {
		var created *models.User
		mockUserRepo.On("CreateUser", mock.AnythingOfType("*models.User")).
			Return(func(u *models.User) *models.User {
				u.ID = uuid.New()
				created = u
				return u
			}, nil).Once()

		user, err := service.CreateDevUser("Onboard Me", nil)
		assert.NoError(t, err)
		assert.Same(t, created, user)
		assert.Nil(t, user.ShowdownUsername)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInvalidInput when name is empty", func(t *testing.T) {
		user, err := service.CreateDevUser("", nil)
		assert.ErrorIs(t, err, types.ErrInvalidInput)
		assert.Nil(t, user)
	})

	t.Run("Returns ErrConflict on duplicate key", func(t *testing.T) {
		mockUserRepo.On("CreateUser", mock.AnythingOfType("*models.User")).
			Return((*models.User)(nil), gorm.ErrDuplicatedKey).Once()

		user, err := service.CreateDevUser("Test User", nil)
		assert.ErrorIs(t, err, types.ErrConflict)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestDevAuthService_Impersonate(t *testing.T) {
	mockUserRepo := new(mock_repositories.MockUserRepository)
	service := newDevAuthService(mockUserRepo, nil, nil, nil)

	userID := uuid.New()

	t.Run("Successfully mints a valid token for an existing user", func(t *testing.T) {
		user := &models.User{ID: userID, DiscordUsername: "alice"}
		mockUserRepo.On("GetUserByID", userID).Return(user, nil).Once()

		returnedUser, token, err := service.Impersonate(userID)
		assert.NoError(t, err)
		assert.Same(t, user, returnedUser)
		assert.NotEmpty(t, token)

		parsedUserID, parseErr := services.NewJWTService("test-secret").ValidateToken(token)
		assert.NoError(t, parseErr)
		assert.Equal(t, userID, parsedUserID)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrUserNotFound if user not found", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", userID).Return((*models.User)(nil), gorm.ErrRecordNotFound).Once()

		user, token, err := service.Impersonate(userID)
		assert.ErrorIs(t, err, types.ErrUserNotFound)
		assert.Nil(t, user)
		assert.Empty(t, token)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService for other repository errors", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", userID).Return((*models.User)(nil), errors.New("db error")).Once()

		user, token, err := service.Impersonate(userID)
		assert.ErrorIs(t, err, types.ErrInternalService)
		assert.Nil(t, user)
		assert.Empty(t, token)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestDevAuthService_UpsertMembership(t *testing.T) {
	leagueID := uuid.New()
	userID := uuid.New()
	user := &models.User{ID: userID, DiscordUsername: "alice"}

	// fresh mocks per subtest so .Once()/.Maybe() expectations never leak
	setup := func() (services.DevAuthService, *mock_repositories.MockLeagueRepository, *mock_repositories.MockUserRepository, *mock_repositories.MockPlayerRepository, *mock_repositories.MockSpectatorRepository) {
		mockUserRepo := new(mock_repositories.MockUserRepository)
		mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
		mockPlayerRepo := new(mock_repositories.MockPlayerRepository)
		mockSpectatorRepo := new(mock_repositories.MockSpectatorRepository)
		service := newDevAuthService(mockUserRepo, mockLeagueRepo, mockPlayerRepo, mockSpectatorRepo)
		return service, mockLeagueRepo, mockUserRepo, mockPlayerRepo, mockSpectatorRepo
	}

	t.Run("Creates player membership when none exists", func(t *testing.T) {
		service, mockLeagueRepo, mockUserRepo, mockPlayerRepo, _ := setup()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(&models.League{ID: leagueID}, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return(user, nil).Once()

		created := &models.Player{}
		var createdInput *models.Player
		mockPlayerRepo.On("GetByUserAndLeague", userID, leagueID).
			Return((*models.Player)(nil), gorm.ErrRecordNotFound).Once()
		mockPlayerRepo.On("Create", mock.AnythingOfType("*models.Player")).
			Return(created, nil).
			Run(func(args mock.Arguments) {
				createdInput = args.Get(0).(*models.Player)
				*created = *createdInput
				created.ID = uuid.New()
			}).Once()

		result, err := service.UpsertMembership(leagueID, userID, "MODERATOR")
		assert.NoError(t, err)
		player, ok := result.(*models.Player)
		require.True(t, ok)
		assert.Same(t, created, player)
		assert.Equal(t, userID, player.UserID)
		assert.Equal(t, leagueID, player.LeagueID)
		assert.Equal(t, rbac.MRoleModerator, player.Role)
		require.NotNil(t, player.TeamName)
		assert.Contains(t, *player.TeamName, "alice")
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Updates role when membership already exists", func(t *testing.T) {
		service, mockLeagueRepo, mockUserRepo, mockPlayerRepo, _ := setup()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(&models.League{ID: leagueID}, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return(user, nil).Once()

		existing := &models.Player{ID: uuid.New(), UserID: userID, LeagueID: leagueID, Role: rbac.MRoleRegular}
		mockPlayerRepo.On("GetByUserAndLeague", userID, leagueID).Return(existing, nil).Once()
		mockPlayerRepo.On("UpdateRole", existing.ID, rbac.MRoleOwner).Return(nil).Once()

		result, err := service.UpsertMembership(leagueID, userID, "OWNER")
		assert.NoError(t, err)
		player, ok := result.(*models.Player)
		require.True(t, ok)
		assert.Same(t, existing, player)
		assert.Equal(t, rbac.MRoleOwner, player.Role)
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("Creates spectator membership for spectator role", func(t *testing.T) {
		service, mockLeagueRepo, mockUserRepo, _, mockSpectatorRepo := setup()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(&models.League{ID: leagueID}, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return(user, nil).Once()

		created := &models.Spectator{}
		mockSpectatorRepo.On("GetByUserAndLeague", userID, leagueID).
			Return((*models.Spectator)(nil), gorm.ErrRecordNotFound).Once()
		mockSpectatorRepo.On("Create", mock.AnythingOfType("*models.Spectator")).
			Return(created, nil).
			Run(func(args mock.Arguments) {
				in := args.Get(0).(*models.Spectator)
				*created = *in
				created.ID = uuid.New()
			}).Once()

		result, err := service.UpsertMembership(leagueID, userID, "SPECTATOR")
		assert.NoError(t, err)
		spectator, ok := result.(*models.Spectator)
		require.True(t, ok)
		assert.Same(t, created, spectator)
		assert.Equal(t, rbac.MRoleSpectator, spectator.Role)
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockSpectatorRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInvalidInput for unknown role", func(t *testing.T) {
		service, _, _, _, _ := setup()

		result, err := service.UpsertMembership(leagueID, userID, "SUPREME_LEADER")
		assert.ErrorIs(t, err, types.ErrInvalidInput)
		assert.Nil(t, result)
	})

	t.Run("Returns ErrLeagueNotFound when league missing", func(t *testing.T) {
		service, mockLeagueRepo, _, _, _ := setup()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return((*models.League)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpsertMembership(leagueID, userID, "REGULAR")
		assert.ErrorIs(t, err, types.ErrLeagueNotFound)
		assert.Nil(t, result)
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrUserNotFound when user missing", func(t *testing.T) {
		service, mockLeagueRepo, mockUserRepo, _, _ := setup()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(&models.League{ID: leagueID}, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return((*models.User)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpsertMembership(leagueID, userID, "REGULAR")
		assert.ErrorIs(t, err, types.ErrUserNotFound)
		assert.Nil(t, result)
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}
