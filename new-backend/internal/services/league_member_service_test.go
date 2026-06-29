package services_test

import (
	"errors"
	"testing"

	mock_repos "github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupLeagueMemberServiceTest() (services.LeagueMemberService, *mock_repos.MockLeagueMemberRepository, *mock_repos.MockLeagueRepository, *mock_repos.MockUserRepository) {
	mockMemberRepo := new(mock_repos.MockLeagueMemberRepository)
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockUserRepo := new(mock_repos.MockUserRepository)

	service := services.NewLeagueMemberService(
		mockMemberRepo,
		mockLeagueRepo,
		mockUserRepo,
	)

	return service, mockMemberRepo, mockLeagueRepo, mockUserRepo
}

func TestLeagueMemberService_GetByID(t *testing.T) {
	service, mockMemberRepo, _, _ := setupLeagueMemberServiceTest()

	t.Run("Success", func(t *testing.T) {
		expected := &models.LeagueMember{ID: uuid.New()}
		mockMemberRepo.On("GetByID", expected.ID).Return(expected, nil).Once()

		result, err := service.GetByID(expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockMemberRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := uuid.New()
		mockMemberRepo.On("GetByID", id).Return((*models.LeagueMember)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetByID(id)
		assert.Error(t, err)
		assert.Equal(t, types.ErrPlayerNotFound, err)
		assert.Nil(t, result)
		mockMemberRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		id := uuid.New()
		mockMemberRepo.On("GetByID", id).Return((*models.LeagueMember)(nil), errors.New("db error")).Once()

		result, err := service.GetByID(id)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, types.ErrInternalService))
		assert.Nil(t, result)
		mockMemberRepo.AssertExpectations(t)
	})
}

func TestLeagueMemberService_GetByUserAndLeague(t *testing.T) {
	service, mockMemberRepo, _, _ := setupLeagueMemberServiceTest()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		leagueID := uuid.New()
		expected := &models.LeagueMember{ID: uuid.New(), UserID: userID, LeagueID: leagueID}
		mockMemberRepo.On("GetByUserAndLeague", userID, leagueID).Return(expected, nil).Once()

		result, err := service.GetByUserAndLeague(userID, leagueID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockMemberRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockMemberRepo.On("GetByUserAndLeague", mock.Anything, mock.Anything).Return((*models.LeagueMember)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetByUserAndLeague(uuid.New(), uuid.New())
		assert.Error(t, err)
		assert.Equal(t, types.ErrPlayerNotFound, err)
		assert.Nil(t, result)
		mockMemberRepo.AssertExpectations(t)
	})
}

func TestLeagueMemberService_GetByLeague(t *testing.T) {
	service, mockMemberRepo, _, _ := setupLeagueMemberServiceTest()

	t.Run("Success", func(t *testing.T) {
		leagueID := uuid.New()
		expected := []models.LeagueMember{{ID: uuid.New(), LeagueID: leagueID}}
		mockMemberRepo.On("GetByLeague", leagueID).Return(expected, nil).Once()

		result, err := service.GetByLeague(leagueID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockMemberRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockMemberRepo.On("GetByLeague", mock.Anything).Return([]models.LeagueMember(nil), errors.New("db error")).Once()

		result, err := service.GetByLeague(uuid.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, types.ErrInternalService))
		assert.Nil(t, result)
		mockMemberRepo.AssertExpectations(t)
	})
}

func TestLeagueMemberService_GetByUser(t *testing.T) {
	service, mockMemberRepo, _, _ := setupLeagueMemberServiceTest()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		expected := []models.LeagueMember{{ID: uuid.New(), UserID: userID}}
		mockMemberRepo.On("GetByUser", userID).Return(expected, nil).Once()

		result, err := service.GetByUser(userID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockMemberRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockMemberRepo.On("GetByUser", mock.Anything).Return([]models.LeagueMember(nil), errors.New("db error")).Once()

		result, err := service.GetByUser(uuid.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, types.ErrInternalService))
		assert.Nil(t, result)
		mockMemberRepo.AssertExpectations(t)
	})
}

func TestLeagueMemberService_GetWithFullRoster(t *testing.T) {
	service, mockMemberRepo, _, _ := setupLeagueMemberServiceTest()

	t.Run("Success", func(t *testing.T) {
		expected := &models.LeagueMember{ID: uuid.New()}
		mockMemberRepo.On("GetWithFullRoster", expected.ID).Return(expected, nil).Once()

		result, err := service.GetWithFullRoster(expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockMemberRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := uuid.New()
		mockMemberRepo.On("GetWithFullRoster", id).Return((*models.LeagueMember)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetWithFullRoster(id)
		assert.Error(t, err)
		assert.Equal(t, types.ErrPlayerNotFound, err)
		assert.Nil(t, result)
		mockMemberRepo.AssertExpectations(t)
	})
}

func TestLeagueMemberService_UpdateProfile(t *testing.T) {
	service, mockMemberRepo, _, _ := setupLeagueMemberServiceTest()

	t.Run("Success", func(t *testing.T) {
		memberID := uuid.New()
		memberUserID := uuid.New()
		leagueID := uuid.New()
		existing := &models.LeagueMember{ID: memberID, UserID: memberUserID, LeagueID: leagueID}

		mockMemberRepo.On("GetByID", memberID).Return(existing, nil).Once()

		result, err := service.UpdateProfile(&models.User{ID: memberUserID}, memberID, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, existing, result)
		mockMemberRepo.AssertExpectations(t)
	})

	t.Run("MemberNotFound", func(t *testing.T) {
		mockMemberRepo.On("GetByID", mock.Anything).Return((*models.LeagueMember)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpdateProfile(&models.User{ID: uuid.New()}, uuid.New(), nil, nil)
		assert.Error(t, err)
		assert.Equal(t, types.ErrPlayerNotFound, err)
		assert.Nil(t, result)
		mockMemberRepo.AssertExpectations(t)
	})
}

func TestLeagueMemberService_UpdateDraftPoints(t *testing.T) {
	service, mockMemberRepo, _, _ := setupLeagueMemberServiceTest()

	t.Run("Success", func(t *testing.T) {
		memberID := uuid.New()
		points := 200
		expected := &models.LeagueMember{ID: memberID, DraftPoints: points}

		mockMemberRepo.On("GetByID", memberID).Return(&models.LeagueMember{ID: memberID}, nil).Once()
		mockMemberRepo.On("GetByID", memberID).Return(expected, nil).Once()
		mockMemberRepo.On("UpdateDraftPoints", memberID, points).Return(nil).Once()

		result, err := service.UpdateDraftPoints(&models.User{ID: uuid.New(), Role: "admin"}, memberID, &points)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockMemberRepo.AssertExpectations(t)
	})
}

func TestLeagueMemberService_UpdateRecord(t *testing.T) {
	service, mockMemberRepo, _, _ := setupLeagueMemberServiceTest()

	t.Run("Success", func(t *testing.T) {
		memberID := uuid.New()
		expected := &models.LeagueMember{ID: memberID, Wins: 5, Losses: 3}

		mockMemberRepo.On("GetByID", memberID).Return(&models.LeagueMember{ID: memberID}, nil).Once()
		mockMemberRepo.On("UpdateRecord", memberID, 5, 3).Return(nil).Once()
		mockMemberRepo.On("GetByID", memberID).Return(expected, nil).Once()

		result, err := service.UpdateRecord(&models.User{ID: uuid.New(), Role: "admin"}, memberID, 5, 3)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockMemberRepo.AssertExpectations(t)
	})
}

func TestLeagueMemberService_UpdateDraftPosition(t *testing.T) {
	service, mockMemberRepo, _, _ := setupLeagueMemberServiceTest()

	t.Run("Success", func(t *testing.T) {
		memberID := uuid.New()
		expected := &models.LeagueMember{ID: memberID, DraftPosition: 3}

		mockMemberRepo.On("GetByID", memberID).Return(&models.LeagueMember{ID: memberID}, nil).Once()
		mockMemberRepo.On("UpdateDraftPosition", memberID, 3).Return(nil).Once()
		mockMemberRepo.On("GetByID", memberID).Return(expected, nil).Once()

		result, err := service.UpdateDraftPosition(&models.User{ID: uuid.New(), Role: "admin"}, memberID, 3)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockMemberRepo.AssertExpectations(t)
	})

	t.Run("MemberNotFound", func(t *testing.T) {
		mockMemberRepo.On("GetByID", mock.Anything).Return((*models.LeagueMember)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpdateDraftPosition(&models.User{ID: uuid.New()}, uuid.New(), 3)
		assert.Error(t, err)
		assert.Equal(t, types.ErrPlayerNotFound, err)
		assert.Nil(t, result)
		mockMemberRepo.AssertExpectations(t)
	})
}
