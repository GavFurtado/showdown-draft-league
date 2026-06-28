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

type DraftPickService interface {
	GetByID(id uuid.UUID) (*models.DraftPick, error)
	GetByDraft(draftID uuid.UUID) ([]models.DraftPick, error)
	GetByPlayer(playerID uuid.UUID) ([]models.DraftPick, error)
	GetCountByDraft(draftID uuid.UUID) (int64, error)
	GetHistory(leagueID uuid.UUID) ([]models.DraftPick, error)
	GetNextPickNumber(draftID uuid.UUID) (int, error)
}

type draftPickServiceImpl struct {
	draftPickRepo repositories.DraftPickRepository
	draftRepo     repositories.DraftRepository
}

func NewDraftPickService(
	draftPickRepo repositories.DraftPickRepository,
	draftRepo repositories.DraftRepository,
) DraftPickService {
	return &draftPickServiceImpl{
		draftPickRepo: draftPickRepo,
		draftRepo:     draftRepo,
	}
}

func (s *draftPickServiceImpl) GetByID(id uuid.UUID) (*models.DraftPick, error) {
	pick, err := s.draftPickRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrDraftPickNotFound
		}
		log.Printf("(Service: DraftPickService.GetByID) - failed: %v\n", err)
		return nil, types.ErrInternalService
	}
	return pick, nil
}

func (s *draftPickServiceImpl) GetByDraft(draftID uuid.UUID) ([]models.DraftPick, error) {
	picks, err := s.draftPickRepo.GetByDraft(draftID)
	if err != nil {
		log.Printf("(Service: DraftPickService.GetByDraft) - failed: %v\n", err)
		return nil, types.ErrInternalService
	}
	return picks, nil
}

func (s *draftPickServiceImpl) GetByPlayer(playerID uuid.UUID) ([]models.DraftPick, error) {
	picks, err := s.draftPickRepo.GetByPlayer(playerID)
	if err != nil {
		log.Printf("(Service: DraftPickService.GetByPlayer) - failed: %v\n", err)
		return nil, types.ErrInternalService
	}
	return picks, nil
}

func (s *draftPickServiceImpl) GetCountByDraft(draftID uuid.UUID) (int64, error) {
	count, err := s.draftPickRepo.GetCountByDraft(draftID)
	if err != nil {
		log.Printf("(Service: DraftPickService.GetCountByDraft) - failed: %v\n", err)
		return 0, types.ErrInternalService
	}
	return count, nil
}

func (s *draftPickServiceImpl) GetHistory(leagueID uuid.UUID) ([]models.DraftPick, error) {
	draft, err := s.draftRepo.GetDraftByLeagueID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrDraftNotFound
		}
		return nil, types.ErrInternalService
	}
	return s.draftPickRepo.GetByDraft(draft.ID)
}

func (s *draftPickServiceImpl) GetNextPickNumber(draftID uuid.UUID) (int, error) {
	count, err := s.draftPickRepo.GetCountByDraft(draftID)
	if err != nil {
		return 0, types.ErrInternalService
	}
	return int(count) + 1, nil
}
