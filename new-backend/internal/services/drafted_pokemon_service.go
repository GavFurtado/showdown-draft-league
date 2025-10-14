package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DraftedPokemonService defines the interface for managing drafted Pokemon.
type DraftedPokemonService interface {
	// gets drafted Pokemon by ID with relationships.
	GetDraftedPokemonByID(id uuid.UUID) (*models.DraftedPokemon, error)
	// gets all Pokemon drafted by a specific player.
	GetDraftedPokemonByPlayer(playerID uuid.UUID) ([]models.DraftedPokemon, error)
	// gets all Pokemon drafted in a specific league.
	GetDraftedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// gets all active (non-released) Pokemon drafted in a league.
	GetActiveDraftedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// gets all released Pokemon (free agents) in a league.
	GetReleasedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// checks if a Pokemon species has been drafted in a league and is not released.
	IsPokemonDrafted(leagueID uuid.UUID, pokemonSpeciesID int64) (bool, error)
	// gets the next draft pick number for a league.
	GetNextDraftPickNumber(leagueID uuid.UUID) (int, error)
	// releases a Pokemon back to free agents.
	ReleasePokemon(currentUser *models.User, draftedPokemonID uuid.UUID) error
	// gets count of actively drafted Pokemon by a player.
	GetDraftedPokemonCountByPlayer(currentUser *models.User, playerID uuid.UUID) (int64, error)
	// gets draft history for a league (all picks in order, including released).
	GetDraftHistory(leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// trades a Pokemon from one player to another.
	TradePokemon(currentUser *models.User, draftedPokemonID, newPlayerID uuid.UUID) error
	// soft deletes a drafted Pokemon entry.
	DeleteDraftedPokemon(currentUser *models.User, draftedPokemonID uuid.UUID) error
	DropPokemon(currentUser *models.User, draftedPokemonID uuid.UUID) error
	PickupFreeAgent(currentUser *models.User, leaguePokemonID uuid.UUID) error
}

type draftedPokemonServiceImpl struct {
	draftedPokemonRepo repositories.DraftedPokemonRepository
	pokemonSpeciesRepo repositories.PokemonSpeciesRepository
	leaguePokemonRepo  repositories.LeaguePokemonRepository
	userRepo           repositories.UserRepository
	leagueRepo         repositories.LeagueRepository
	playerRepo         repositories.PlayerRepository
}

// NewDraftedPokemonService creates a new instance of DraftedPokemonService.
func NewDraftedPokemonService(
	draftedPokemonRepo repositories.DraftedPokemonRepository,
	userRepo repositories.UserRepository,
	leagueRepo repositories.LeagueRepository,
	playerRepo repositories.PlayerRepository,
	pokemonSpeciesRepo repositories.PokemonSpeciesRepository,
	leaguePokemonRepo repositories.LeaguePokemonRepository,
) DraftedPokemonService {
	return &draftedPokemonServiceImpl{
		draftedPokemonRepo: draftedPokemonRepo,
		userRepo:           userRepo,
		leagueRepo:         leagueRepo,
		playerRepo:         playerRepo,
		pokemonSpeciesRepo: pokemonSpeciesRepo,
		leaguePokemonRepo:  leaguePokemonRepo,
	}
}

// isLeagueInTransferWindow checks if the specified league is currently in a transfer window.
func (s *draftedPokemonServiceImpl) isLeagueInTransferWindow(leagueID uuid.UUID) (bool, error) {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		return false, common.ErrLeagueNotFound
	}
	return league.Status == enums.LeagueStatusTransferWindow, nil
}

// DropPokemon allows a user to drop a drafted Pokemon, making it a free agent.
// This operation is only allowed during a transfer window and if the user owns the Pokemon.
func (s *draftedPokemonServiceImpl) DropPokemon(currentUser *models.User, draftedPokemonID uuid.UUID) error {
	draftedPokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(draftedPokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrDraftedPokemonNotFound
		}
		return common.ErrInternalService
	}

	// Authorize user is owner of pokemon
	player, err := s.playerRepo.GetPlayerByID(draftedPokemon.PlayerID)
	if err != nil {
		return common.ErrPlayerNotFound
	}
	if player.UserID != currentUser.ID {
		return common.ErrUnauthorized
	}

	inWindow, err := s.isLeagueInTransferWindow(draftedPokemon.LeagueID)
	if err != nil {
		return err
	}
	if !inWindow {
		return common.ErrInvalidState // Or a more specific "not in transfer window" error
	}

	return s.ReleasePokemon(currentUser, draftedPokemonID)
}

// PickupFreeAgent allows a user to pick up a released Pokemon (free agent) using transfer credits.
// This operation is only allowed during a transfer window and if the player has enough credits.
func (s *draftedPokemonServiceImpl) PickupFreeAgent(currentUser *models.User, leaguePokemonID uuid.UUID) error {
	leaguePokemon, err := s.leaguePokemonRepo.GetLeaguePokemonByID(leaguePokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrLeaguePokemonNotFound
		}
		return common.ErrInternalService
	}

	inWindow, err := s.isLeagueInTransferWindow(leaguePokemon.LeagueID)
	if err != nil {
		return err
	}
	if !inWindow {
		return common.ErrInvalidState // Or a more specific "not in transfer window" error
	}

	player, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, leaguePokemon.LeagueID)
	if err != nil {
		return common.ErrPlayerNotFound
	}

	if player.TransferCredits <= 0 {
		return common.ErrInsufficientTransferCredits
	}

	if !leaguePokemon.IsAvailable {
		return common.ErrConflict // Pokemon not available
	}

	newDraftedPokemon := &models.DraftedPokemon{
		LeagueID:         leaguePokemon.LeagueID,
		PlayerID:         player.ID,
		PokemonSpeciesID: leaguePokemon.PokemonSpeciesID,
		LeaguePokemonID:  leaguePokemon.ID,
		IsReleased:       false,
	}

	// Call the new repository transaction method
	if err := s.draftedPokemonRepo.PickupFreeAgentTransaction(player, newDraftedPokemon, leaguePokemon); err != nil {
		log.Printf("LOG: (Error: DraftedPokemonService.PickupFreeAgent) - Failed to complete pickup free agent transaction: %v", err)
		return common.ErrInternalService
	}

	return nil
}

// GetDraftedPokemonByID gets drafted Pokemon by ID with relationships.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonByID(id uuid.UUID) (*models.DraftedPokemon, error) {
	pokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrDraftedPokemonNotFound
		}
		log.Printf("LOG: (Error: DraftedPokemonService.GetDraftedPokemonByID) - Failed to get drafted pokemon by ID %s: %v", id, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// GetDraftedPokemonByPlayer gets all Pokemon drafted by a specific player.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonByPlayer(playerID uuid.UUID) ([]models.DraftedPokemon, error) {
	targetPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("LOG: (Error: DraftedPokemonService.GetDraftedPokemonByPlayer) - Failed to get target player %s: %v", playerID, err)
		return nil, common.ErrInternalService
	}

	pokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByPlayer(targetPlayer.ID)
	if err != nil {
		log.Printf("LOG: (Error: DraftedPokemonService.GetDraftedPokemonByPlayer) - Failed to get drafted pokemon by player %s: %v", playerID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// GetDraftedPokemonByLeague gets all Pokemon drafted in a specific league.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	pokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonByLeague) - Failed to get drafted pokemon by league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// GetActiveDraftedPokemonByLeague gets all active (non-released) Pokemon drafted in a league.
func (s *draftedPokemonServiceImpl) GetActiveDraftedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	pokemon, err := s.draftedPokemonRepo.GetActiveDraftedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetActiveDraftedPokemonByLeague) - Failed to get active drafted pokemon by league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// GetReleasedPokemonByLeague gets all released Pokemon in a league.
func (s *draftedPokemonServiceImpl) GetReleasedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	pokemon, err := s.draftedPokemonRepo.GetReleasedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetReleasedPokemonByLeague) - Failed to get released pokemon by league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// IsPokemonDrafted checks if a Pokemon species has been drafted in a league and is not released.
func (s *draftedPokemonServiceImpl) IsPokemonDrafted(leagueID uuid.UUID, pokemonSpeciesID int64) (bool, error) {
	// check if valid species id
	if _, err := s.pokemonSpeciesRepo.GetPokemonSpeciesByID(pokemonSpeciesID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, common.ErrPokemonSpeciesNotFound
		}
		return false, common.ErrInternalService
	}

	isDrafted, err := s.draftedPokemonRepo.IsPokemonDrafted(leagueID, pokemonSpeciesID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.IsPokemonDrafted) - Failed to check if pokemon is drafted for league %s and species %d: %v\n", leagueID, pokemonSpeciesID, err)
		return false, common.ErrInternalService
	}

	return isDrafted, nil
}

// GetNextDraftPickNumber gets the next draft pick number for a league.
func (s *draftedPokemonServiceImpl) GetNextDraftPickNumber(leagueID uuid.UUID) (int, error) {
	leagueStatus, err := s.leagueRepo.GetLeagueStatus(leagueID)
	if err != nil {
		log.Printf("LOG: (Error: DraftedPokemonService.GetNextDraftPickNumber) - Failed to get league status for league %s: %v", leagueID, err)
		return 0, common.ErrInternalService
	}

	if leagueStatus != enums.LeagueStatusDrafting {
		return 0, common.ErrInvalidState
	}

	nextPick, err := s.draftedPokemonRepo.GetNextDraftPickNumber(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetNextDraftPickNumber) - Failed to get next draft pick number for league %s: %v", leagueID, err)
		return 0, common.ErrInternalService
	}

	return nextPick, nil
}

// ReleasePokemon releases a Pokemon back to free agents.
func (s *draftedPokemonServiceImpl) ReleasePokemon(currentUser *models.User, draftedPokemonID uuid.UUID) error {
	draftedPokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(draftedPokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrDraftedPokemonNotFound
		}
		log.Printf("LOG: (Error: DraftedPokemonService.ReleasePokemon) - Error getting drafted pokemon %s for release: %v", draftedPokemonID, err)
		return common.ErrInternalService
	}

	if draftedPokemon.IsReleased {
		return common.ErrPokemonAlreadyReleased
	}

	// Get the player who owns this pokemon to check authorization
	// otPlayer = Original Trainer Player. renamed from ownerPlayer to prevent confusion with player role owner
	otPlayer, err := s.playerRepo.GetPlayerByID(draftedPokemon.PlayerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrPlayerNotFound
		}
		log.Printf("LOG: (Error: DraftedPokemonService.ReleasePokemon) - Error getting owner player %s for drafted pokemon %s: %v", draftedPokemon.PlayerID, draftedPokemonID, err)
		return common.ErrInternalService
	}

	if currentUser.Role != "admin" {
		currentPlayer, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, draftedPokemon.LeagueID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("LOG: (Error: DraftedPokemonService.ReleasePokemon) - No player found for user %s in league %s: %v\n", currentUser.ID, draftedPokemon.LeagueID, err)
				return common.ErrPlayerNotFound
			}
			log.Printf("LOG: (Error: DraftedPokemonService.ReleasePokemon) - Error fetching current player (league: %d) of user %s: %v\n", draftedPokemon.LeagueID, currentUser.ID, err)
			return common.ErrInternalService
		}

		if currentUser.ID != otPlayer.UserID && currentPlayer.Role == rbac.PRoleMember {
			log.Printf("LOG: (Error: DraftedPokemonService.ReleasePokemon) - Unauthorized attempt by user %s to release pokemon %s", currentUser.ID, draftedPokemonID)
			return common.ErrUnauthorized
		}
	}

	// If we reach this point, the user is authorized to release the pokemon.
	err = s.draftedPokemonRepo.ReleasePokemon(draftedPokemonID)
	if err != nil {
		log.Printf("LOG: (Error: DraftedPokemonService.ReleasePokemon) - Failed to release pokemon with ID %s: %v", draftedPokemonID, err)
		return common.ErrInternalService
	}
	return nil
}

// GetDraftedPokemonCountByPlayer gets count of actively drafted Pokemon by a player.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonCountByPlayer(currentUser *models.User, playerID uuid.UUID) (int64, error) {
	count, err := s.draftedPokemonRepo.GetDraftedPokemonCountByPlayer(playerID)
	if err != nil {
		log.Printf("LOG: (Error: DraftedPokemonService.GetDraftedPokemonCountByPlayer) - (user %s) Failed to get drafted pokemon count for player %s: %v", currentUser.ID, playerID, err)
		return 0, common.ErrInternalService
	}

	return count, nil
}

// GetDraftHistory gets draft history for a league (all picks in order, including released and includes transfers).
func (s *draftedPokemonServiceImpl) GetDraftHistory(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	history, err := s.draftedPokemonRepo.GetDraftHistory(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetDraftHistory) - Failed to get draft history for league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return history, nil
}

// TradePokemon trades a Pokemon from one player to another.
// TODO: this is very basic for now. what we want is a full blown trade offer system. (not planned for anytime soon)
func (s *draftedPokemonServiceImpl) TradePokemon(currentUser *models.User, draftedPokemonID, newPlayerID uuid.UUID) error {
	// 1. Get the drafted Pokemon details for authorization and validation
	draftedPokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(draftedPokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrDraftedPokemonNotFound
		}
		log.Printf("(Error: TradePokemon) - Error getting drafted pokemon %s for trade: %v", draftedPokemonID, err)
		return common.ErrInternalService
	}

	// 2. Get the current owner of the pokemon
	currentOwnerPlayer, err := s.playerRepo.GetPlayerByID(draftedPokemon.PlayerID)
	if err != nil {
		log.Printf("(Error: TradePokemon) - Error getting current owner player %s for drafted pokemon %s: %v", draftedPokemon.PlayerID, draftedPokemonID, err)
		return common.ErrInternalService
	}

	// 3. Get the new player's details
	newPlayer, err := s.playerRepo.GetPlayerByID(newPlayerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("new player for trade not found")
		}
		log.Printf("(Error: TradePokemon) - Failed to get new player %s: %v", newPlayerID, err)
		return fmt.Errorf("failed to get new player for trade: %w", err)
	}

	// Authorization: Admin, Owner of the league, or the current owner of the Pokemon.
	isOwner, err := s.leagueRepo.IsUserOwner(currentUser.ID, draftedPokemon.LeagueID)
	if err != nil {
		return err
	}

	// Basic authorization: Admin, Owner, or the current owner can initiate/approve.
	// More complex trade logic (e.g., both players agree) would be implemented here or in a higher-level "Trade" service.
	if currentUser.Role != "admin" && !isOwner && currentUser.ID != currentOwnerPlayer.UserID {
		log.Printf("(Error: DraftedPokemonService.TradePokemon) - Unauthorized attempt by user %s to trade pokemon %s", currentUser.ID, draftedPokemonID)
		return common.ErrUnauthorized
	}

	// Validation: Ensure the pokemon is not released
	if draftedPokemon.IsReleased {
		return errors.New("released pokemon cannot be traded")
	}
	// Validation: Ensure both players are in the same league (as the pokemon)
	if currentOwnerPlayer.LeagueID != newPlayer.LeagueID || currentOwnerPlayer.LeagueID != draftedPokemon.LeagueID {
		return errors.New("trade involves players or pokemon from different leagues")
	}

	err = s.draftedPokemonRepo.TradePokemon(draftedPokemonID, newPlayerID)
	if err != nil {
		log.Printf("(Error: TradePokemon) - Failed to trade pokemon with ID %s to player %s: %v", draftedPokemonID, newPlayerID, err)
		return fmt.Errorf("failed to trade pokemon: %w", err)
	}

	return nil
}

// DeleteDraftedPokemon soft deletes a drafted Pokemon entry.
// Player permission required: rbac.PermissionDeleteDraftedPokemon
func (s *draftedPokemonServiceImpl) DeleteDraftedPokemon(currentUser *models.User, draftedPokemonID uuid.UUID) error {
	err := s.draftedPokemonRepo.DeleteDraftedPokemon(draftedPokemonID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("LOG: (Error: DraftedPokemonService.DeleteDraftedPokemon) - drafted pokemon %s not found: %v", draftedPokemonID, err)
			return common.ErrDraftedPokemonNotFound
		}
		log.Printf("(Error: DraftedPokemonService.DeleteDraftedPokemon) - Failed to delete drafted pokemon with ID %s: %v", draftedPokemonID, err)
		return common.ErrInternalService
	}

	return nil
}

