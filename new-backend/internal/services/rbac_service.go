package services

import (
	"errors"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RBACService defines the interface for Role-Based Access Control operations.
type RBACService interface {
	CanAccess(userID uuid.UUID, leagueID uuid.UUID, requiredPermission rbac.Permission) (*models.Player, bool, error)
}

type RBACServiceImpl struct {
	leagueRepo repositories.LeagueRepository
	userRepo   repositories.UserRepository
	playerRepo repositories.PlayerRepository
}

// NewRBACService creates a new instance of RBACService.
func NewRBACService(leagueRepo repositories.LeagueRepository, userRepo repositories.UserRepository, playerRepo repositories.PlayerRepository) RBACService {
	return &RBACServiceImpl{
		leagueRepo: leagueRepo,
		userRepo:   userRepo,
		playerRepo: playerRepo,
	}
}

// CanAccess checks if a user has the required permission for a specific action.
func (s *RBACServiceImpl) CanAccess(userID uuid.UUID, leagueID uuid.UUID, requiredPermission rbac.Permission) (*models.Player, bool, error) {
	player, err := s.playerRepo.GetPlayerByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("LOG: (Service: CanAccess) - Player (User ID: %s) not found (likely not part of league or league doesn't exist).\n", userID)
			return nil, false, types.ErrLeagueNotFound
		}
		log.Printf("LOG: (Service: CanAccess) - failed to retrieve player (userID: %s; leagueID: %s\n", userID, leagueID)
		return nil, false, types.ErrInternalService
	}

	// check if the role matches the permission required
	return player, player.Can(requiredPermission), nil
}
