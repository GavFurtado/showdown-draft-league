package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// defines the interface for user-related business logic.
type UserService interface {
	GetMyProfileHandler(userID uuid.UUID) (*models.User, error)
	GetMyDiscordDetailsHandler(userID uuid.UUID) (*common.DiscordUser, error)
	UpdateProfileHandler(userID uuid.UUID, req common.UpdateProfileRequest) (*models.User, error)
}

type userServiceImpl struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
	}
}

// retrieves the full user profile.
func (s *userServiceImpl) GetMyProfileHandler(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		log.Printf("(Error: GetMyProfileHandler) - Failed to get user %s from repository: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve user profile: %w", err)
	}
	return user, nil
}

// retrieves formatted Discord-specific user details.
func (s *userServiceImpl) GetMyDiscordDetailsHandler(userID uuid.UUID) (*common.DiscordUser, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		log.Printf("(Error: GetMyDiscordDetailsHandler) - Failed to get user %s from repository: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve Discord details: %w", err)
	}

	discordDeets := common.DiscordUser{
		ID:       user.ID.String(),
		Username: user.DiscordUsername,
		Avatar:   user.DiscordAvatarURL,
	}

	return &discordDeets, nil
}

// updates profile with request fields
func (s *userServiceImpl) UpdateProfileHandler(userID uuid.UUID, input common.UpdateProfileRequest) (*models.User, error) {
	log.Printf("UserService: UpdateProfileHandler called for user %s with request: %+v", userID, input)

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		log.Printf("(Error: UpdateProfileHandler) - User fetch failed: %s", err.Error())
		return nil, err
	}

	if input.ShowdownName != "" {
		user.ShowdownUsername = input.ShowdownName
	}

	updatedUser, err := s.userRepo.UpdateUser(user)
	if err != nil {
		log.Printf("(Error: UpdateProfileHandler) - Update failed: %v", err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}
