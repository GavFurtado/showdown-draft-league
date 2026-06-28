package services

import (
	"errors"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
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
}

func NewPoolEntryService(
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
	}
}

func (s *poolEntryServiceImpl) getLeagueByID(leagueID, currentUserID uuid.UUID) (*models.League, error) {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Service: PoolEntryService.getLeagueByID) - could not find league %s. (currentUser.ID: %s)\n", leagueID, currentUserID)
			return nil, types.ErrLeagueNotFound
		}
		log.Printf("(Service: PoolEntryService.getLeagueByID) - could not retrieve league by leagueID %s (currentUser.ID: %s)\n", leagueID, currentUserID)
		return nil, types.ErrInternalService
	}
	return league, nil
}

func (s *poolEntryServiceImpl) getPokemonSpeciesByID(pokemonSpeciesID int64) (*models.PokemonSpecies, error) {
	pokemon, err := s.pokemonSpeciesRepo.GetPokemonSpeciesByID(pokemonSpeciesID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Service: PoolEntryService.getPokemonSpeciesByID) - pokemon %d species not found: %v\n", pokemonSpeciesID, err)
			return nil, types.ErrPokemonSpeciesNotFound
		}
		log.Printf("(Service: PoolEntryService.getPokemonSpeciesByID) - could not retrieve pokemon species %d.\n", pokemonSpeciesID)
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
		log.Printf("(Service: PoolEntryService.GetByID) - failed: %v\n", err)
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
		log.Printf("(Service: PoolEntryService.GetByLeague) - failed: %v\n", err)
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
		log.Printf("(Service: PoolEntryService.GetAvailableByLeague) - failed: %v\n", err)
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
		log.Printf("LOG: (Service: PoolEntryService.Create) - operation not allowed for current league status: %s for user %s", league.Status, currentUser.ID)
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
		log.Printf("LOG: (Service: PoolEntryService.Create) - failed: %v\n", err)
		return nil, types.ErrInternalService
	}

	log.Printf("LOG: (Service: PoolEntryService.Create) - Successfully created pool entry for league %s, species %d", input.LeagueID, input.PokemonSpeciesID)
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
			log.Printf("LOG: (Service: PoolEntryService.CreateBatch) - operation not allowed for current league status: %s for user %s", league.Status, currentUser.ID)
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
		log.Printf("LOG: (Service: PoolEntryService.CreateBatch) - failed: %v\n", err)
		return nil, types.ErrInternalService
	}

	log.Printf("LOG: (Service: PoolEntryService.CreateBatch) - Successfully batch created %d pool entries", len(created))
	return created, nil
}

func (s *poolEntryServiceImpl) Update(currentUser *models.User, input *requests.PoolEntryUpdateRequestDTO) (*models.PoolEntry, error) {
	existing, err := s.poolEntryRepo.GetByID(input.PoolEntryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Service: PoolEntryService.Update) - pool entry %s does not exist: %v\n", input.PoolEntryID, err)
			return nil, types.ErrPoolEntryNotFound
		}
		log.Printf("(Service: PoolEntryService.Update) - could not fetch pool entry: %s\n", err.Error())
		return nil, types.ErrInternalService
	}

	league, err := s.leagueRepo.GetLeagueByID(existing.LeagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Service: PoolEntryService.Update) - league %s does not exist: %s\n", existing.LeagueID, err.Error())
			return nil, types.ErrLeagueNotFound
		}
		log.Printf("(Service: PoolEntryService.Update) - could not fetch league %s: %v\n", existing.LeagueID, err)
		return nil, types.ErrInternalService
	}

	if currentUser.Role != "admin" &&
		(league.Status != enums.LeagueStatusSetup && league.Status != enums.LeagueStatusDrafting) {
		log.Printf("(Service: PoolEntryService.Update) - operation not allowed for current league status: %s for user %s", league.Status, currentUser.ID)
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
		log.Printf("(Service: PoolEntryService.Update) - failed: %s\n", err.Error())
		return nil, types.ErrInternalService
	}

	log.Printf("(Service: PoolEntryService.Update) - Successfully updated pool entry %s", updated.ID)
	return updated, nil
}
