package services

import (
	"errors"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// defines the interface for Role-Based Access Control operations.
type RBACService interface {
	CanAccess(userID uuid.UUID, leagueID uuid.UUID, requiredPermission rbac.Permission) (bool, error)
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

// checks if a user has the required permission for a specific action.
func (s *RBACServiceImpl) CanAccess(userID uuid.UUID, leagueID uuid.UUID, requiredPermission rbac.Permission) (bool, error) {
	player, err := s.playerRepo.GetPlayerByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Service: CanAccess) - Player (User ID: %s) not found (likely not part of league or league doesn't exist).\n", userID)
			return false, common.ErrLeagueNotFound
		}
		log.Printf("(Service: CanAccess) - failed to retrieve player (userID: %s; leagueID: %s\n", userID, leagueID)
		return false, common.ErrInternalService
	}

	return player.Can(rbac.Permission(requiredPermission)), nil
}
