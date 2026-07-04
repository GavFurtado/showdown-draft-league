package services

import (
	"errors"
	"log/slog"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PoolEntryService interface {
	GetByID(id uuid.UUID) (*models.PoolEntry, error)
	GetByLeague(leagueID uuid.UUID) ([]models.PoolEntry, error)
	GetAvailableByLeague(leagueID uuid.UUID) ([]models.PoolEntry, error)
	Create(currentUser *models.User, input *requests.PoolEntryCreateRequestDTO) (*models.PoolEntry, error)
	CreateBatch(currentUser *models.User, inputs []requests.PoolEntryCreateRequestDTO) ([]models.PoolEntry, error)
	Update(currentUser *models.User, input *requests.PoolEntryUpdateRequestDTO) (*models.PoolEntry, error)
}

type poolEntryServiceImpl struct {
	poolEntryRepo      repositories.PoolEntryRepository
	leagueRepo         repositories.LeagueRepository
	userRepo           repositories.UserRepository
	pokemonSpeciesRepo repositories.PokemonSpeciesRepository
	logger             *slog.Logger
}

func NewPoolEntryService(
	logger *slog.Logger,
	poolEntryRepo repositories.PoolEntryRepository,
	leagueRepo repositories.LeagueRepository,
	userRepo repositories.UserRepository,
	pokemonSpeciesRepo repositories.PokemonSpeciesRepository,
) PoolEntryService {
	return &poolEntryServiceImpl{
		poolEntryRepo:      poolEntryRepo,
		leagueRepo:         leagueRepo,
		userRepo:           userRepo,
		pokemonSpeciesRepo: pokemonSpeciesRepo,
		logger:             utils.LoggerWithService(logger, "PoolEntryService"),
	}
}

func (s *poolEntryServiceImpl) getLeagueByID(leagueID, currentUserID uuid.UUID) (*models.League, error) {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("getLeagueByID - league not found", "league_id", leagueID, "current_user_id", currentUserID)
			return nil, types.ErrLeagueNotFound
		}
		s.logger.Error("getLeagueByID - could not retrieve league", "league_id", leagueID, "current_user_id", currentUserID, "error", err)
		return nil, types.ErrInternalService
	}
	return league, nil
}

func (s *poolEntryServiceImpl) getPokemonSpeciesByID(pokemonSpeciesID int64) (*models.PokemonSpecies, error) {
	pokemon, err := s.pokemonSpeciesRepo.GetPokemonSpeciesByID(pokemonSpeciesID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("getPokemonSpeciesByID - pokemon species not found", "species_id", pokemonSpeciesID, "error", err)
			return nil, types.ErrPokemonSpeciesNotFound
		}
		s.logger.Error("getPokemonSpeciesByID - could not retrieve pokemon species", "species_id", pokemonSpeciesID, "error", err)
		return nil, types.ErrInternalService
	}
	return pokemon, nil
}

func (s *poolEntryServiceImpl) GetByID(id uuid.UUID) (*models.PoolEntry, error) {
	entry, err := s.poolEntryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPoolEntryNotFound
		}
		s.logger.Error("GetByID - failed", "error", err)
		return nil, types.ErrInternalService
	}
	return entry, nil
}

func (s *poolEntryServiceImpl) GetByLeague(leagueID uuid.UUID) ([]models.PoolEntry, error) {
	entries, err := s.poolEntryRepo.GetByLeague(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrLeagueNotFound
		}
		s.logger.Error("GetByLeague - failed", "error", err)
		return nil, types.ErrInternalService
	}
	return entries, nil
}

func (s *poolEntryServiceImpl) GetAvailableByLeague(leagueID uuid.UUID) ([]models.PoolEntry, error) {
	entries, err := s.poolEntryRepo.GetAvailableByLeague(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrLeagueNotFound
		}
		s.logger.Error("GetAvailableByLeague - failed", "error", err)
		return nil, types.ErrInternalService
	}
	return entries, nil
}

func (s *poolEntryServiceImpl) Create(currentUser *models.User, input *requests.PoolEntryCreateRequestDTO) (*models.PoolEntry, error) {
	league, err := s.getLeagueByID(input.LeagueID, currentUser.ID)
	if err != nil {
		return nil, err
	}

	if league.Status != enums.LeagueStatusSetup {
		s.logger.Warn("Create - operation not allowed for current league status", "league_status", league.Status, "user_id", currentUser.ID)
		return nil, types.ErrInvalidState
	}

	_, err = s.getPokemonSpeciesByID(input.PokemonSpeciesID)
	if err != nil {
		return nil, err
	}

	entry := &models.PoolEntry{
		LeagueID:         input.LeagueID,
		PokemonSpeciesID: input.PokemonSpeciesID,
		Cost:             input.Cost,
		IsAvailable:      true,
	}

	created, err := s.poolEntryRepo.Create(entry)
	if err != nil {
		s.logger.Error("Create - failed", "error", err)
		return nil, types.ErrInternalService
	}

	s.logger.Info("Create - successfully created pool entry", "league_id", input.LeagueID, "species_id", input.PokemonSpeciesID)
	return created, nil
}

func (s *poolEntryServiceImpl) CreateBatch(currentUser *models.User, inputs []requests.PoolEntryCreateRequestDTO) ([]models.PoolEntry, error) {
	if len(inputs) == 0 {
		return []models.PoolEntry{}, nil
	}

	leagueCache := make(map[uuid.UUID]*models.League)
	var entriesToCreate []models.PoolEntry

	for _, input := range inputs {
		league, exists := leagueCache[input.LeagueID]
		if !exists {
			var err error
			league, err = s.getLeagueByID(input.LeagueID, currentUser.ID)
			if err != nil {
				return nil, err
			}
			leagueCache[input.LeagueID] = league
		}

		if league.Status != enums.LeagueStatusSetup {
			s.logger.Warn("CreateBatch - operation not allowed for current league status", "league_status", league.Status, "user_id", currentUser.ID)
			return nil, types.ErrInvalidState
		}

		_, err := s.getPokemonSpeciesByID(input.PokemonSpeciesID)
		if err != nil {
			return nil, err
		}

		entriesToCreate = append(entriesToCreate, models.PoolEntry{
			LeagueID:         input.LeagueID,
			PokemonSpeciesID: input.PokemonSpeciesID,
			Cost:             input.Cost,
			IsAvailable:      true,
		})
	}

	created, err := s.poolEntryRepo.CreateBatch(entriesToCreate)
	if err != nil {
		s.logger.Error("CreateBatch - failed", "error", err)
		return nil, types.ErrInternalService
	}

	s.logger.Info("CreateBatch - successfully batch created pool entries", "count", len(created))
	return created, nil
}

func (s *poolEntryServiceImpl) Update(currentUser *models.User, input *requests.PoolEntryUpdateRequestDTO) (*models.PoolEntry, error) {
	existing, err := s.poolEntryRepo.GetByID(input.PoolEntryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Update - pool entry does not exist", "pool_entry_id", input.PoolEntryID, "error", err)
			return nil, types.ErrPoolEntryNotFound
		}
		s.logger.Error("Update - could not fetch pool entry", "error", err.Error())
		return nil, types.ErrInternalService
	}

	league, err := s.leagueRepo.GetLeagueByID(existing.LeagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Update - league does not exist", "league_id", existing.LeagueID, "error", err.Error())
			return nil, types.ErrLeagueNotFound
		}
		s.logger.Error("Update - could not fetch league", "league_id", existing.LeagueID, "error", err)
		return nil, types.ErrInternalService
	}

	if currentUser.Role != "admin" &&
		(league.Status != enums.LeagueStatusSetup && league.Status != enums.LeagueStatusDrafting) {
		s.logger.Warn("Update - operation not allowed for current league status", "league_status", league.Status, "user_id", currentUser.ID)
		return nil, types.ErrInvalidState
	}

	if input.Cost != nil && *input.Cost != *existing.Cost {
		existing.Cost = input.Cost
	}
	if *input.IsAvailable != existing.IsAvailable {
		existing.IsAvailable = *input.IsAvailable
	}

	updated, err := s.poolEntryRepo.Update(existing)
	if err != nil {
		s.logger.Error("Update - failed", "error", err.Error())
		return nil, types.ErrInternalService
	}

	s.logger.Info("Update - successfully updated pool entry", "pool_entry_id", updated.ID)
	return updated, nil
}
