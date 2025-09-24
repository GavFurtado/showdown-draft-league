package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DraftedPokemonService interface {
	// gets drafted Pokemon by ID with relationships.
	GetDraftedPokemonByID(currentUser *models.User, id uuid.UUID) (*models.DraftedPokemon, error)
	// gets all Pokemon drafted by a specific player.
	GetDraftedPokemonByPlayer(currentUser *models.User, playerID uuid.UUID) ([]models.DraftedPokemon, error)
	// gets all Pokemon drafted in a specific league.
	GetDraftedPokemonByLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// gets all active (non-released) Pokemon drafted in a league.
	GetActiveDraftedPokemonByLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// gets all released Pokemon (free agents) in a league.
	GetReleasedPokemonByLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// checks if a Pokemon species has been drafted in a league and is not released.
	IsPokemonDrafted(currentUser *models.User, leagueID, pokemonSpeciesID uuid.UUID) (bool, error)
	// gets the next draft pick number for a league.
	GetNextDraftPickNumber(currentUser *models.User, leagueID uuid.UUID) (int, error)
	// releases a Pokemon back to free agents.
	ReleasePokemon(currentUser *models.User, draftedPokemonID uuid.UUID) error
	// gets count of actively drafted Pokemon by a player.
	GetDraftedPokemonCountByPlayer(currentUser *models.User, playerID uuid.UUID) (int64, error)
	// gets draft history for a league (all picks in order, including released).
	GetDraftHistory(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// trades a Pokemon from one player to another.
	TradePokemon(currentUser *models.User, draftedPokemonID, newPlayerID uuid.UUID) error
	// performs a draft transaction (create DraftedPokemon and update LeaguePokemon availability).
	DraftPokemonTransaction(
		currentUser *models.User,
		draftedPokemon *models.DraftedPokemon,
		leagueID, pokemonSpeciesID uuid.UUID,
	) error
	// soft deletes a drafted Pokemon entry.
	DeleteDraftedPokemon(currentUser *models.User, draftedPokemonID uuid.UUID) error
}

type draftedPokemonServiceImpl struct {
	draftedPokemonRepo repositories.DraftedPokemonRepository
	userRepo           repositories.UserRepository
	leagueRepo         repositories.LeagueRepository
	playerRepo         repositories.PlayerRepository
}

func NewDraftedPokemonService(
	draftedPokemonRepo repositories.DraftedPokemonRepository,
	userRepo repositories.UserRepository,
	leagueRepo repositories.LeagueRepository,
	playerRepo repositories.PlayerRepository,
) DraftedPokemonService {
	return &draftedPokemonServiceImpl{
		draftedPokemonRepo: draftedPokemonRepo,
		userRepo:           userRepo,
		leagueRepo:         leagueRepo,
		playerRepo:         playerRepo,
	}
}

// gets drafted Pokemon by ID with relationships.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonByID(currentUser *models.User, id uuid.UUID) (*models.DraftedPokemon, error) {
	pokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrPokemonSpeciesNotFound
		}
		log.Printf("LOG: (Error: DraftedPokemonService.GetDraftedPokemonByID) - Failed to get drafted pokemon by ID %s: %v", id, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// gets all Pokemon drafted by a specific player.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonByPlayer(currentUser *models.User, playerID uuid.UUID) ([]models.DraftedPokemon, error) {
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

// gets all Pokemon drafted in a specific league.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonByLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	pokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonByLeague) - Failed to get drafted pokemon by league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// gets all active (non-released) Pokemon drafted in a league.
func (s *draftedPokemonServiceImpl) GetActiveDraftedPokemonByLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	pokemon, err := s.draftedPokemonRepo.GetActiveDraftedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetActiveDraftedPokemonByLeague) - Failed to get active drafted pokemon by league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// gets all released Pokemon in a league.
func (s *draftedPokemonServiceImpl) GetReleasedPokemonByLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	pokemon, err := s.draftedPokemonRepo.GetReleasedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetReleasedPokemonByLeague) - Failed to get released pokemon by league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// checks if a Pokemon species has been drafted in a league and is not released.
func (s *draftedPokemonServiceImpl) IsPokemonDrafted(currentUser *models.User, leagueID, pokemonSpeciesID uuid.UUID) (bool, error) {
	isDrafted, err := s.draftedPokemonRepo.IsPokemonDrafted(leagueID, pokemonSpeciesID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.IsPokemonDrafted) - Failed to check if pokemon is drafted for league %s and species %s: %v", leagueID, pokemonSpeciesID, err)
		return false, common.ErrInternalService
	}

	return isDrafted, nil
}

// gets the next draft pick number for a league.
func (s *draftedPokemonServiceImpl) GetNextDraftPickNumber(currentUser *models.User, leagueID uuid.UUID) (int, error) {
	nextPick, err := s.draftedPokemonRepo.GetNextDraftPickNumber(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetNextDraftPickNumber) - Failed to get next draft pick number for league %s: %v", leagueID, err)
		return 0, common.ErrInternalService
	}

	return nextPick, nil
}

// releases a Pokemon back to free agents.
func (s *draftedPokemonServiceImpl) ReleasePokemon(currentUser *models.User, draftedPokemonID uuid.UUID) error {
	draftedPokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(draftedPokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrDraftedPokemonNotFound
		}
		log.Printf("(Error: DraftedPokemonService.ReleasePokemon) - Error getting drafted pokemon %s for release: %v", draftedPokemonID, err)
		return common.ErrInternalService
	}

	if draftedPokemon.IsReleased {
		return errors.New("pokemon is already released")
	}

	// Get the player who owns this pokemon to check authorization
	// otPlayer = Original Trainer Player. renamed from ownerPlayer to prevent confusion with player role owner
	otPlayer, err := s.playerRepo.GetPlayerByID(draftedPokemon.PlayerID)
	if err != nil {
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
			log.Printf("(Error: DraftedPokemonService.ReleasePokemon) - Unauthorized attempt by user %s to release pokemon %s", currentUser.ID, draftedPokemonID)
			return common.ErrUnauthorized
		}
	}

	err = s.draftedPokemonRepo.ReleasePokemon(draftedPokemonID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.ReleasePokemon) - Failed to release pokemon with ID %s: %v", draftedPokemonID, err)
		return common.ErrInternalService
	}
	return nil
}

// gets count of actively drafted Pokemon by a player.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonCountByPlayer(currentUser *models.User, playerID uuid.UUID) (int64, error) {
	count, err := s.draftedPokemonRepo.GetDraftedPokemonCountByPlayer(playerID)
	if err != nil {
		log.Printf("LOG: (Error: DraftedPokemonService.GetDraftedPokemonCountByPlayer) - (user %s) Failed to get drafted pokemon count for player %s: %v", currentUser.ID, playerID, err)
		return 0, common.ErrInternalService
	}

	return count, nil
}

// gets draft history for a league (all picks in order, including released).
func (s *draftedPokemonServiceImpl) GetDraftHistory(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	history, err := s.draftedPokemonRepo.GetDraftHistory(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetDraftHistory) - Failed to get draft history for league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return history, nil
}

// TODO: this is very basic for now. what we want is a full blown trade offer system. (not planned for anytime soon)
// trades a Pokemon from one player to another.
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
		log.Printf("(Error: DraftedPokemonService.TradePokemon) - Failed to trade pokemon with ID %s to player %s: %v", draftedPokemonID, newPlayerID, err)
		return fmt.Errorf("failed to trade pokemon: %w", err)
	}

	return nil
}

// performs a draft transaction (create DraftedPokemon and update LeaguePokemon availability).
func (s *draftedPokemonServiceImpl) DraftPokemonTransaction(
	currentUser *models.User,
	draftedPokemon *models.DraftedPokemon,
	leagueID, pokemonSpeciesID uuid.UUID,
) error {
	// Authorization: Admin, Owner of the league, or the player making their own draft pick.
	isOwner, err := s.leagueRepo.IsUserOwner(currentUser.ID, leagueID)
	if err != nil {
		return err
	}

	// Get the player entity for the current user in this specific league.
	// needed to verify if the current user is the player making the pick.
	currentPlayer, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User is not a player, cannot draft unless admin/owner
			if currentUser.Role != "admin" && !isOwner {
				return common.ErrUnauthorized
			}
		} else {
			log.Printf("(Error: DraftPokemonTransaction) - Failed to get player for user %s in league %s: %v", currentUser.ID, leagueID, err)
			return common.ErrInternalService
		}
	}

	// If not admin and not owner, ensure the user is drafting for themselves.
	if currentUser.Role != "admin" && !isOwner {
		if currentPlayer == nil || currentPlayer.ID != draftedPokemon.PlayerID {
			log.Printf("(Error: DraftPokemonTransaction) - Unauthorized attempt by user %s to draft for player %s in league %s (not self/admin/owner)", currentUser.ID, draftedPokemon.PlayerID, leagueID)
			return common.ErrUnauthorized
		}
	}
	// More specific draft-turn-based authorization logic would go here.
	// check if it's actually `currentPlayer`'s turn.

	// Add more specific validation/checks before starting the transaction, e.g.,
	// Does the league exist? (Implicitly checked by owner/player check)
	// Does the pokemon species exist in the league pool and is it available? (Can be checked here)
	// Does the player have enough draft points? (Can be checked here)

	err = s.draftedPokemonRepo.DraftPokemonTransaction(draftedPokemon)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.DraftPokemonTransaction) - Failed to perform draft transaction for league %s, species %s: %v", leagueID, pokemonSpeciesID, err)
		return fmt.Errorf("failed to perform draft transaction: %w", err)
	}

	return nil
}

// soft deletes a drafted Pokemon entry.
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
