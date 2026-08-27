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

func newDevAuthService(mockUserRepo *mock_repositories.MockUserRepository, mockLeagueRepo *mock_repositories.MockLeagueRepository, mockMemberRepo *mock_repositories.MockLeagueMemberRepository) services.DevAuthService {
	jwtService := services.NewJWTService("test-secret")
	return services.NewDevAuthService(slog.Default(), mockUserRepo, mockLeagueRepo, mockMemberRepo, jwtService)
}

func TestDevAuthService_ListUsers(t *testing.T) {
	mockUserRepo := new(mock_repositories.MockUserRepository)
	service := newDevAuthService(mockUserRepo, nil, nil)

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
	service := newDevAuthService(mockUserRepo, nil, nil)

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
	service := newDevAuthService(mockUserRepo, nil, nil)

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
	setup := func() (services.DevAuthService, *mock_repositories.MockLeagueRepository, *mock_repositories.MockUserRepository, *mock_repositories.MockLeagueMemberRepository) {
		mockUserRepo := new(mock_repositories.MockUserRepository)
		mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
		mockMemberRepo := new(mock_repositories.MockLeagueMemberRepository)
		service := newDevAuthService(mockUserRepo, mockLeagueRepo, mockMemberRepo)
		return service, mockLeagueRepo, mockUserRepo, mockMemberRepo
	}

	t.Run("Creates membership when none exists", func(t *testing.T) {
		service, mockLeagueRepo, mockUserRepo, mockMemberRepo := setup()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(&models.League{ID: leagueID}, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return(user, nil).Once()

		created := &models.LeagueMember{}
		var createdInput *models.LeagueMember
		mockMemberRepo.On("GetByUserAndLeague", userID, leagueID).
			Return((*models.LeagueMember)(nil), gorm.ErrRecordNotFound).Once()
		mockMemberRepo.On("Create", mock.AnythingOfType("*models.LeagueMember")).
			Return(created, nil).
			Run(func(args mock.Arguments) {
				createdInput = args.Get(0).(*models.LeagueMember)
				*created = *createdInput
				created.ID = uuid.New()
			}).Once()

		member, err := service.UpsertMembership(leagueID, userID, "MODERATOR")
		assert.NoError(t, err)
		assert.Same(t, created, member)
		assert.Equal(t, userID, member.UserID)
		assert.Equal(t, leagueID, member.LeagueID)
		assert.Equal(t, rbac.MRoleModerator, member.Role)
		assert.True(t, member.IsParticipating)
		require.NotNil(t, member.TeamName)
		assert.Contains(t, *member.TeamName, "alice")
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockMemberRepo.AssertExpectations(t)
	})

	t.Run("Updates role when membership already exists", func(t *testing.T) {
		service, mockLeagueRepo, mockUserRepo, mockMemberRepo := setup()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(&models.League{ID: leagueID}, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return(user, nil).Once()

		existing := &models.LeagueMember{ID: uuid.New(), UserID: userID, LeagueID: leagueID, Role: rbac.MRoleMember}
		mockMemberRepo.On("GetByUserAndLeague", userID, leagueID).Return(existing, nil).Once()
		mockMemberRepo.On("UpdateRole", existing.ID, rbac.MRoleOwner).Return(nil).Once()

		member, err := service.UpsertMembership(leagueID, userID, "OWNER")
		assert.NoError(t, err)
		assert.Same(t, existing, member)
		assert.Equal(t, rbac.MRoleOwner, member.Role)
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockMemberRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInvalidInput for unknown role", func(t *testing.T) {
		service, _, _, _ := setup()

		member, err := service.UpsertMembership(leagueID, userID, "SUPREME_LEADER")
		assert.ErrorIs(t, err, types.ErrInvalidInput)
		assert.Nil(t, member)
	})

	t.Run("Returns ErrLeagueNotFound when league missing", func(t *testing.T) {
		service, mockLeagueRepo, _, _ := setup()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return((*models.League)(nil), gorm.ErrRecordNotFound).Once()

		member, err := service.UpsertMembership(leagueID, userID, "MEMBER")
		assert.ErrorIs(t, err, types.ErrLeagueNotFound)
		assert.Nil(t, member)
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrUserNotFound when user missing", func(t *testing.T) {
		service, mockLeagueRepo, mockUserRepo, _ := setup()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(&models.League{ID: leagueID}, nil).Once()
		mockUserRepo.On("GetUserByID", userID).Return((*models.User)(nil), gorm.ErrRecordNotFound).Once()

		member, err := service.UpsertMembership(leagueID, userID, "MEMBER")
		assert.ErrorIs(t, err, types.ErrUserNotFound)
		assert.Nil(t, member)
		mockLeagueRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}
