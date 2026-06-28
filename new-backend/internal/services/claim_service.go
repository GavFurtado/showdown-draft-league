package services

import (
	"errors"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
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
}

func NewClaimService(claimRepo repositories.ClaimRepository) ClaimService {
	return &claimServiceImpl{claimRepo: claimRepo}
}

func (s *claimServiceImpl) GetByID(id uuid.UUID) (*models.Claim, error) {
	claim, err := s.claimRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrClaimNotFound
		}
		log.Printf("(Service: ClaimService.GetByID) - failed: %v\n", err)
		return nil, types.ErrInternalService
	}
	return claim, nil
}

func (s *claimServiceImpl) GetActiveByPlayer(playerID uuid.UUID) ([]models.Claim, error) {
	claims, err := s.claimRepo.GetActiveByPlayer(playerID)
	if err != nil {
		log.Printf("(Service: ClaimService.GetActiveByPlayer) - failed: %v\n", err)
		return nil, types.ErrInternalService
	}
	return claims, nil
}

func (s *claimServiceImpl) GetActiveByLeague(leagueID uuid.UUID) ([]models.Claim, error) {
	claims, err := s.claimRepo.GetActiveByLeague(leagueID)
	if err != nil {
		log.Printf("(Service: ClaimService.GetActiveByLeague) - failed: %v\n", err)
		return nil, types.ErrInternalService
	}
	return claims, nil
}

func (s *claimServiceImpl) GetReleasedByLeague(leagueID uuid.UUID) ([]models.Claim, error) {
	claims, err := s.claimRepo.GetReleasedByLeague(leagueID)
	if err != nil {
		log.Printf("(Service: ClaimService.GetReleasedByLeague) - failed: %v\n", err)
		return nil, types.ErrInternalService
	}
	return claims, nil
}
