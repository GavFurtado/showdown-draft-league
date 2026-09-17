package services

import (
	"errors"
	"log/slog"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RBACService defines the interface for Role-Based Access Control operations.
type RBACService interface {
	CanAccess(userID uuid.UUID, leagueID uuid.UUID, requiredPermission rbac.Permission) (bool, error)
}

type RBACServiceImpl struct {
	leagueRepo    repositories.LeagueRepository
	userRepo      repositories.UserRepository
	playerRepo    repositories.PlayerRepository
	spectatorRepo repositories.SpectatorRepository
	logger        *slog.Logger
}

// NewRBACService creates a new instance of RBACService.
func NewRBACService(logger *slog.Logger, leagueRepo repositories.LeagueRepository, userRepo repositories.UserRepository, playerRepo repositories.PlayerRepository, spectatorRepo repositories.SpectatorRepository) RBACService {
	return &RBACServiceImpl{
		leagueRepo:    leagueRepo,
		userRepo:      userRepo,
		playerRepo:    playerRepo,
		spectatorRepo: spectatorRepo,
		logger:        utils.LoggerWithService(logger, "RBACService"),
	}
}

// CanAccess checks if a user has the required permission for a specific action.
// The user's MemberRole is resolved from either their Player or Spectator row —
// role is independent of participation. Spectator roles are derived from their
// player counterparts minus participation permissions, so no runtime
// participation filter is needed: the role table alone decides.
func (s *RBACServiceImpl) CanAccess(userID uuid.UUID, leagueID uuid.UUID, requiredPermission rbac.Permission) (bool, error) {
	player, err := s.playerRepo.GetByUserAndLeague(userID, leagueID)
	if err == nil {
		return player.Can(requiredPermission), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("failed to retrieve player", "error", err, "user_id", userID, "league_id", leagueID)
		return false, types.ErrInternalService
	}

	spectator, err := s.spectatorRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("membership not found - likely not part of league or league doesn't exist", "user_id", userID, "league_id", leagueID)
			return false, types.ErrLeagueNotFound
		}
		s.logger.Error("failed to retrieve spectator", "error", err, "user_id", userID, "league_id", leagueID)
		return false, types.ErrInternalService
	}

	if spectator.IsExpired() {
		return false, nil
	}
	if spectator.IsViewer {
		return rbac.IsReadPermission(requiredPermission), nil
	}
	return spectator.Can(requiredPermission), nil
}
