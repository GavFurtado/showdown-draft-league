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
	CanAccess(userID uuid.UUID, leagueID uuid.UUID, requiredPermission rbac.Permission) (*models.LeagueMember, bool, error)
}

type RBACServiceImpl struct {
	leagueRepo repositories.LeagueRepository
	userRepo   repositories.UserRepository
	memberRepo repositories.LeagueMemberRepository
}

// NewRBACService creates a new instance of RBACService.
func NewRBACService(leagueRepo repositories.LeagueRepository, userRepo repositories.UserRepository, memberRepo repositories.LeagueMemberRepository) RBACService {
	return &RBACServiceImpl{
		leagueRepo: leagueRepo,
		userRepo:   userRepo,
		memberRepo: memberRepo,
	}
}

// CanAccess checks if a user has the required permission for a specific action.
func (s *RBACServiceImpl) CanAccess(userID uuid.UUID, leagueID uuid.UUID, requiredPermission rbac.Permission) (*models.LeagueMember, bool, error) {
	member, err := s.memberRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("LOG: (Service: CanAccess) - Member (User ID: %s) not found (likely not part of league or league doesn't exist).\n", userID)
			return nil, false, types.ErrLeagueNotFound
		}
		log.Printf("LOG: (Service: CanAccess) - failed to retrieve member (userID: %s; leagueID: %s\n", userID, leagueID)
		return nil, false, types.ErrInternalService
	}

	return member, member.Can(requiredPermission), nil
}
