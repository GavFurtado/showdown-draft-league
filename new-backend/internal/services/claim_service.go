package services

import (
	"errors"
	"log/slog"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClaimService interface {
	GetByID(id uuid.UUID) (*models.Claim, error)
	GetActiveByPlayer(playerID uuid.UUID) ([]models.Claim, error)
	GetActiveByLeague(leagueID uuid.UUID) ([]models.Claim, error)
	GetReleasedByLeague(leagueID uuid.UUID) ([]models.Claim, error)
}

type claimServiceImpl struct {
	claimRepo repositories.ClaimRepository
	logger    *slog.Logger
}

func NewClaimService(logger *slog.Logger, claimRepo repositories.ClaimRepository) ClaimService {
	return &claimServiceImpl{
		claimRepo: claimRepo,
		logger:    utils.LoggerWithService(logger, "ClaimService"),
	}
}

func (s *claimServiceImpl) GetByID(id uuid.UUID) (*models.Claim, error) {
	claim, err := s.claimRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrClaimNotFound
		}
		s.logger.Error("GetByID - failed", "error", err)
		return nil, types.ErrInternalService
	}
	return claim, nil
}

func (s *claimServiceImpl) GetActiveByPlayer(playerID uuid.UUID) ([]models.Claim, error) {
	claims, err := s.claimRepo.GetActiveByPlayer(playerID)
	if err != nil {
		s.logger.Error("GetActiveByPlayer - failed", "error", err)
		return nil, types.ErrInternalService
	}
	return claims, nil
}

func (s *claimServiceImpl) GetActiveByLeague(leagueID uuid.UUID) ([]models.Claim, error) {
	claims, err := s.claimRepo.GetActiveByLeague(leagueID)
	if err != nil {
		s.logger.Error("GetActiveByLeague - failed", "error", err)
		return nil, types.ErrInternalService
	}
	return claims, nil
}

func (s *claimServiceImpl) GetReleasedByLeague(leagueID uuid.UUID) ([]models.Claim, error) {
	claims, err := s.claimRepo.GetReleasedByLeague(leagueID)
	if err != nil {
		s.logger.Error("GetReleasedByLeague - failed", "error", err)
		return nil, types.ErrInternalService
	}
	return claims, nil
}
