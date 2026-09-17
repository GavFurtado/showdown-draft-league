package services

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SpectatorService interface {
	GetByLeague(leagueID uuid.UUID) ([]models.Spectator, error)
	JoinLeague(currentUserID, leagueID uuid.UUID, inLeagueName *string) (*models.Spectator, error)
	LeaveLeague(userID, leagueID uuid.UUID) error
	UpdateProfile(spectatorID uuid.UUID, inLeagueName *string) (*models.Spectator, error)
	UpdateOwnProfile(userID, leagueID uuid.UUID, inLeagueName *string) (*models.Spectator, error)
	UpdateRole(spectatorID uuid.UUID, newRole rbac.MemberRole) (*models.Spectator, error)
	DeleteSpectator(spectatorID uuid.UUID) error
}

type spectatorServiceImpl struct {
	spectatorRepo repositories.SpectatorRepository
	playerRepo    repositories.PlayerRepository
	leagueRepo    repositories.LeagueRepository
	logger        *slog.Logger
}

func NewSpectatorService(
	logger *slog.Logger,
	spectatorRepo repositories.SpectatorRepository,
	playerRepo repositories.PlayerRepository,
	leagueRepo repositories.LeagueRepository,
) SpectatorService {
	return &spectatorServiceImpl{
		spectatorRepo: spectatorRepo,
		playerRepo:    playerRepo,
		leagueRepo:    leagueRepo,
		logger:        utils.LoggerWithService(logger, "SpectatorService"),
	}
}

func (s *spectatorServiceImpl) GetByLeague(leagueID uuid.UUID) ([]models.Spectator, error) {
	spectators, err := s.spectatorRepo.GetByLeague(leagueID)
	if err != nil {
		s.logger.Error("Failed to retrieve spectators", "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve spectators", types.ErrInternalService)
	}
	return spectators, nil
}

func (s *spectatorServiceImpl) JoinLeague(currentUserID, leagueID uuid.UUID, inLeagueName *string) (*models.Spectator, error) {
	if _, err := s.leagueRepo.GetLeagueByID(leagueID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrLeagueNotFound
		}
		s.logger.Error("Failed to fetch league", "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("%w: league lookup failed", types.ErrInternalService)
	}

	existingPlayer, err := s.playerRepo.FindByUserAndLeague(currentUserID, leagueID)
	if err != nil {
		s.logger.Error("Failed to check existing player", "user_id", currentUserID, "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("%w: existing player lookup failed", types.ErrInternalService)
	}
	if existingPlayer != nil {
		return nil, types.ErrUserAlreadyInLeague
	}

	existingSpectator, err := s.spectatorRepo.GetByUserAndLeague(currentUserID, leagueID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("Failed to check existing spectator", "user_id", currentUserID, "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("%w: existing spectator lookup failed", types.ErrInternalService)
	}
	if existingSpectator != nil {
		return nil, types.ErrUserAlreadySpectator
	}

	spectator := models.Spectator{
		UserID:       currentUserID,
		LeagueID:     leagueID,
		InLeagueName: inLeagueName,
		Role:         rbac.MRoleSpectator,
		IsViewer:     false,
	}

	created, err := s.spectatorRepo.Create(&spectator)
	if err != nil {
		s.logger.Error("Failed to create spectator", "user_id", currentUserID, "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("%w: spectator creation failed", types.ErrInternalService)
	}

	s.logger.Info("Spectator joined", "spectator_id", created.ID, "user_id", currentUserID, "league_id", leagueID)
	return created, nil
}

func (s *spectatorServiceImpl) LeaveLeague(userID, leagueID uuid.UUID) error {
	existing, err := s.spectatorRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrSpectatorNotFound
		}
		s.logger.Error("Failed to retrieve spectator", "user_id", userID, "league_id", leagueID, "error", err)
		return fmt.Errorf("%w: leaving league lookup failed", types.ErrInternalService)
	}

	if existing.Role.IsOwner() {
		return types.ErrCannotDemoteLeagueOwner
	}

	if err := s.spectatorRepo.Delete(existing.ID); err != nil {
		s.logger.Error("Failed to delete spectator", "spectator_id", existing.ID, "error", err)
		return fmt.Errorf("%w: failed to leave league", types.ErrInternalService)
	}

	return nil
}

func (s *spectatorServiceImpl) UpdateProfile(spectatorID uuid.UUID, inLeagueName *string) (*models.Spectator, error) {
	existing, err := s.spectatorRepo.GetByID(spectatorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrSpectatorNotFound
		}
		return nil, fmt.Errorf("%w: spectator profile lookup failed", types.ErrInternalService)
	}

	if inLeagueName != nil {
		existing.InLeagueName = inLeagueName
	}

	return s.spectatorRepo.Update(existing)
}

func (s *spectatorServiceImpl) UpdateOwnProfile(userID, leagueID uuid.UUID, inLeagueName *string) (*models.Spectator, error) {
	existing, err := s.spectatorRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrSpectatorNotFound
		}
		return nil, fmt.Errorf("%w: own spectator profile lookup failed", types.ErrInternalService)
	}

	if inLeagueName != nil {
		existing.InLeagueName = inLeagueName
	}

	return s.spectatorRepo.Update(existing)
}

func (s *spectatorServiceImpl) UpdateRole(spectatorID uuid.UUID, newRole rbac.MemberRole) (*models.Spectator, error) {
	// Role transitions stay within the spectator roles; becoming a player is
	// handled separately (promote/demote).
	if newRole.IsParticipant() {
		return nil, types.ErrInvalidInput
	}

	if _, err := s.spectatorRepo.GetByID(spectatorID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrSpectatorNotFound
		}
		return nil, types.ErrInternalService
	}

	if err := s.spectatorRepo.UpdateRole(spectatorID, newRole); err != nil {
		return nil, types.ErrInternalService
	}

	updated, err := s.spectatorRepo.GetByID(spectatorID)
	if err != nil {
		return nil, types.ErrInternalService
	}

	return updated, nil
}

func (s *spectatorServiceImpl) DeleteSpectator(spectatorID uuid.UUID) error {
	existing, err := s.spectatorRepo.GetByID(spectatorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrSpectatorNotFound
		}
		s.logger.Error("Failed to retrieve spectator", "spectator_id", spectatorID, "error", err)
		return fmt.Errorf("%w: spectator removal lookup failed", types.ErrInternalService)
	}

	if existing.Role.IsOwner() {
		return types.ErrCannotDemoteLeagueOwner
	}

	if err := s.spectatorRepo.Delete(spectatorID); err != nil {
		s.logger.Error("Failed to delete spectator", "spectator_id", spectatorID, "error", err)
		return fmt.Errorf("%w: failed to delete spectator", types.ErrInternalService)
	}

	return nil
}
