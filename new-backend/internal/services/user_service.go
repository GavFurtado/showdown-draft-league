package services

import (
	"errors"
	"log/slog"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/responses"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TODO: remove the Handler suffix we have here for the functions
// handlers in go conventions are generally controller

// UserService defines the interface for user-related business logic.
type UserService interface {
	GetMyProfileHandler(userID uuid.UUID) (*models.User, error)
	GetMyDiscordDetailsHandler(userID uuid.UUID) (*responses.DiscordUserResponse, error)
	UpdateProfileHandler(userID uuid.UUID, req requests.UserUpdateProfileRequestDTO) (*models.User, error)
	GetMyLeaguesHandler(userID uuid.UUID) ([]*models.League, error)
}

type userServiceImpl struct {
	userRepo repositories.UserRepository
	logger   *slog.Logger
}

func NewUserService(logger *slog.Logger, userRepo repositories.UserRepository) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
		logger:   utils.LoggerWithService(logger, "UserService"),
	}
}

// retrieves the full user profile.
func (s *userServiceImpl) GetMyProfileHandler(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}
		s.logger.Error("GetMyProfileHandler - failed to get user from repository", "user_id", userID, "error", err)
		return nil, types.ErrInternalService
	}
	return user, nil
}

// retrieves formatted Discord-specific user details.
func (s *userServiceImpl) GetMyDiscordDetailsHandler(userID uuid.UUID) (*responses.DiscordUserResponse, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}
		s.logger.Error("GetMyDiscordDetailsHandler - failed to get user from repository", "user_id", userID, "error", err)
		return nil, types.ErrInternalService
	}

	discordDeets := responses.DiscordUserResponse{
		ID:       user.ID.String(),
		Username: user.DiscordUsername,
		Avatar:   user.DiscordAvatarURL,
	}

	return &discordDeets, nil
}

// updates profile with request fields
func (s *userServiceImpl) UpdateProfileHandler(userID uuid.UUID, input requests.UserUpdateProfileRequestDTO) (*models.User, error) {
	s.logger.Info("UpdateProfileHandler called", "user_id", userID, "request", input)

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}
		s.logger.Error("UpdateProfileHandler - user fetch failed", "error", err.Error())
		return nil, types.ErrInternalService
	}

	if input.ShowdownName != nil {
		user.ShowdownUsername = input.ShowdownName
	}

	updatedUser, err := s.userRepo.UpdateUser(user)
	if err != nil {
		s.logger.Error("UpdateProfileHandler - update failed", "error", err)
		return nil, types.ErrInternalService
	}

	return updatedUser, nil
}

func (s *userServiceImpl) GetMyLeaguesHandler(userID uuid.UUID) ([]*models.League, error) {
	s.logger.Info("GetMyLeaguesHandler called", "user_id", userID)

	leagues, err := s.userRepo.GetUserLeagues(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("user not found when fetching leagues", "user_id", userID)
			return nil, types.ErrUserNotFound
		}
		// other errors
		s.logger.Error("failed to retrieve leagues for user", "user_id", userID, "error", err)
		return nil, types.ErrInternalService
	}

	return leagues, nil
}
