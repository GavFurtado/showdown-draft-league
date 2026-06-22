package services

import (
	"errors"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/responses"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// defines the interface for user-related business logic.
type UserService interface {
	GetMyProfileHandler(userID uuid.UUID) (*models.User, error)
	GetMyDiscordDetailsHandler(userID uuid.UUID) (*responses.DiscordUserResponse, error)
	UpdateProfileHandler(userID uuid.UUID, req requests.UserUpdateProfileRequestDTO) (*models.User, error)
	GetMyLeaguesHandler(userID uuid.UUID) ([]*models.League, error)
}

type userServiceImpl struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
	}
}

// retrieves the full user profile.
func (s *userServiceImpl) GetMyProfileHandler(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}
		log.Printf("(Error: GetMyProfileHandler) - Failed to get user %s from repository: %v", userID, err)
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
		log.Printf("(Error: GetMyDiscordDetailsHandler) - Failed to get user %s from repository: %v", userID, err)
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
	log.Printf("UserService: UpdateProfileHandler called for user %s with request: %+v", userID, input)

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}
		log.Printf("(Error: UpdateProfileHandler) - User fetch failed: %s", err.Error())
		return nil, types.ErrInternalService
	}

	if input.ShowdownName != nil {
		user.ShowdownUsername = *input.ShowdownName
	}

	updatedUser, err := s.userRepo.UpdateUser(user)
	if err != nil {
		log.Printf("(Error: UpdateProfileHandler) - Update failed: %v", err)
		return nil, types.ErrInternalService
	}

	return updatedUser, nil
}

func (s *userServiceImpl) GetMyLeaguesHandler(userID uuid.UUID) ([]*models.League, error) {
	log.Printf("UserService: GetMyLeaguesHandler called for user %s", userID)

	leagues, err := s.userRepo.GetUserLeagues(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("UserService: User %s not found when fetching leagues.", userID)
			return nil, types.ErrUserNotFound
		}
		// other errors
		log.Printf("UserService: Failed to retrieve leagues for user %s: %v", userID, err)
		return nil, types.ErrInternalService
	}

	return leagues, nil
}
