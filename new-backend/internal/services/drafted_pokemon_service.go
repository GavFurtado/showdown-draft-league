package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
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
	// // re-drafts a released Pokemon (from free agents) to a new player.
	// ReDraftPokemon(currentUser *models.User, draftedPokemonID, newPlayerID uuid.UUID, newPickNumber int) error
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
	draftedPokemonRepo *repositories.DraftedPokemonRepository
	userRepo           *repositories.UserRepository
	leagueRepo         *repositories.LeagueRepository
	playerRepo         *repositories.PlayerRepository
}

func NewDraftedPokemonService(
	draftedPokemonRepo *repositories.DraftedPokemonRepository,
	userRepo *repositories.UserRepository,
	leagueRepo *repositories.LeagueRepository,
	playerRepo *repositories.PlayerRepository,
) DraftedPokemonService {
	return &draftedPokemonServiceImpl{
		draftedPokemonRepo: draftedPokemonRepo,
		userRepo:           userRepo,
		leagueRepo:         leagueRepo,
		playerRepo:         playerRepo,
	}
}

// --- Private Helper Authorization Methods ---
// These encapsulate error handling for repository calls
func (s *draftedPokemonServiceImpl) isUserCommissioner(userID, leagueID uuid.UUID) (bool, error) {
	isComm, err := s.leagueRepo.IsUserCommissioner(userID, leagueID)
	if err != nil {
		log.Printf("(Error: draftedPokemonService.isUserCommissioner) - Failed to check commissioner status for user %s in league %s: %v", userID, leagueID, err)
		return false, fmt.Errorf("failed to check commissioner status: %w", err)
	}
	return isComm, nil
}

func (s *draftedPokemonServiceImpl) isUserPlayerInLeague(userID, leagueID uuid.UUID) (bool, error) {
	isPlayer, err := s.leagueRepo.IsUserPlayerInLeague(userID, leagueID)
	if err != nil {
		log.Printf("(Error: draftedPokemonService.isUserPlayerInLeague) - Failed to check player status for user %s in league %s: %v", userID, leagueID, err)
		return false, fmt.Errorf("failed to check player status: %w", err)
	}
	return isPlayer, nil
}

// --- Service Methods ---

// creates a new drafted Pokemon entry.
func (s *draftedPokemonServiceImpl) CreateDraftedPokemon(req *common.DraftedPokemonCreateRequest) (*models.DraftedPokemon, error) {
	draftedPokemon := &models.DraftedPokemon{
		LeagueID:         req.LeagueID,
		PlayerID:         req.PlayerID,
		PokemonSpeciesID: req.PokemonSpeciesID,
		DraftRoundNumber: req.DraftRoundNumber,
		DraftPickNumber:  req.DraftPickNumber,
	}

	if req.IsReleased != nil {
		draftedPokemon.IsReleased = *req.IsReleased
	} else {
		// GORM handles `default:false`.
	}

	// TODO: adjust this once draft model is fixed
	// Perhaps check the Draft model's current Draft number
	// check if drafting is completed etc (although controller shouldn't let a req reach here in that case)
	if draftedPokemon.DraftRoundNumber == 0 {
		// rn 0 is fine
	}
	if draftedPokemon.DraftPickNumber == 0 {
		// 0 is fine for now
	}

	returnedDraftedPokemon, err := s.draftedPokemonRepo.CreateDraftedPokemon(draftedPokemon)
	if err != nil {
		return nil, err
	}

	return returnedDraftedPokemon, nil
}

// gets drafted Pokemon by ID with relationships.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonByID(currentUser *models.User, id uuid.UUID) (*models.DraftedPokemon, error) {
	pokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrPokemonSpeciesNotFound
		}
		log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonByID) - Failed to get drafted pokemon by ID %s: %v", id, err)
		return nil, common.ErrInternalService
	}

	// Authorization check: User must be admin or a player in the league the pokemon belongs to.
	isPlayerInLeague, err := s.isUserPlayerInLeague(currentUser.ID, pokemon.LeagueID)
	if err != nil {
		return nil, err // Error already logged in helper
	}

	if !currentUser.IsAdmin && !isPlayerInLeague { // if player is not in the league and is not an admin
		log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonByID) - Unauthorized access to drafted pokemon %s by user %s", id, currentUser.ID)
		return nil, common.ErrUnauthorized
	}

	return pokemon, nil
}

// gets all Pokemon drafted by a specific player.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonByPlayer(currentUser *models.User, playerID uuid.UUID) ([]models.DraftedPokemon, error) {
	// First, get the target player to check their UserID and LeagueID.
	targetPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonByPlayer) - Failed to get target player %s: %v", playerID, err)
		return nil, common.ErrInternalService
	}

	// Authorization: Admin, or the player themselves, or a player in the same league.
	if !currentUser.IsAdmin {
		if currentUser.ID != targetPlayer.UserID {
			isCurrentUserInTargetLeague, err := s.isUserPlayerInLeague(currentUser.ID, targetPlayer.LeagueID)
			if err != nil {
				return nil, err
			}
			if !isCurrentUserInTargetLeague {
				log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonByPlayer) - Unauthorized access to player %s's drafted pokemon by user %s", playerID, currentUser.ID)
				return nil, common.ErrUnauthorized
			}
		}
	}

	pokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByPlayer(playerID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonByPlayer) - Failed to get drafted pokemon by player %s: %v", playerID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// gets all Pokemon drafted in a specific league.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonByLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	// Authorization check: Admin or a player in the league.
	isPlayerInLeague, err := s.isUserPlayerInLeague(currentUser.ID, leagueID)
	if err != nil {
		return nil, err
	}

	if !currentUser.IsAdmin && !isPlayerInLeague {
		log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonByLeague) - Unauthorized attempt by user %s for league %s", currentUser.ID, leagueID)
		return nil, common.ErrUnauthorized
	}

	pokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonByLeague) - Failed to get drafted pokemon by league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// gets all active (non-released) Pokemon drafted in a league.
func (s *draftedPokemonServiceImpl) GetActiveDraftedPokemonByLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	// Authorization check: Admin or a player in the league.
	isPlayerInLeague, err := s.isUserPlayerInLeague(currentUser.ID, leagueID)
	if err != nil {
		return nil, err
	}

	if !currentUser.IsAdmin && !isPlayerInLeague {
		log.Printf("(Error: DraftedPokemonService.GetActiveDraftedPokemonByLeague) - Unauthorized attempt by user %s for league %s", currentUser.ID, leagueID)
		return nil, common.ErrUnauthorized
	}

	pokemon, err := s.draftedPokemonRepo.GetActiveDraftedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetActiveDraftedPokemonByLeague) - Failed to get active drafted pokemon by league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// gets all released Pokemon (free agents) in a league.
func (s *draftedPokemonServiceImpl) GetReleasedPokemonByLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	// Authorization check: Admin or a player in the league.
	isPlayerInLeague, err := s.isUserPlayerInLeague(currentUser.ID, leagueID)
	if err != nil {
		return nil, err // Error already logged in helper
	}

	if !currentUser.IsAdmin && !isPlayerInLeague {
		log.Printf("(Error: DraftedPokemonService.GetReleasedPokemonByLeague) - Unauthorized attempt by user %s for league %s", currentUser.ID, leagueID)
		return nil, common.ErrUnauthorized
	}

	pokemon, err := s.draftedPokemonRepo.GetReleasedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetReleasedPokemonByLeague) - Failed to get released pokemon by league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return pokemon, nil
}

// checks if a Pokemon species has been drafted in a league and is not released.
func (s *draftedPokemonServiceImpl) IsPokemonDrafted(currentUser *models.User, leagueID, pokemonSpeciesID uuid.UUID) (bool, error) {
	// Authorization check: Admin or a player in the league.
	isPlayerInLeague, err := s.isUserPlayerInLeague(currentUser.ID, leagueID)
	if err != nil {
		return false, err
	}

	if !currentUser.IsAdmin && !isPlayerInLeague {
		log.Printf("(Error: DraftedPokemonService.IsPokemonDrafted) - Unauthorized attempt by user %s for league %s", currentUser.ID, leagueID)
		return false, common.ErrUnauthorized
	}

	isDrafted, err := s.draftedPokemonRepo.IsPokemonDrafted(leagueID, pokemonSpeciesID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.IsPokemonDrafted) - Failed to check if pokemon is drafted for league %s and species %s: %v", leagueID, pokemonSpeciesID, err)
		return false, common.ErrInternalService
	}

	return isDrafted, nil
}

// gets the next draft pick number for a league.
func (s *draftedPokemonServiceImpl) GetNextDraftPickNumber(currentUser *models.User, leagueID uuid.UUID) (int, error) {
	// Authorization check: Admin or a player in the league.
	isPlayerInLeague, err := s.isUserPlayerInLeague(currentUser.ID, leagueID)
	if err != nil {
		return 0, err
	}

	if !currentUser.IsAdmin && !isPlayerInLeague {
		log.Printf("(Error: DraftedPokemonService.GetNextDraftPickNumber) - Unauthorized attempt by user %s for league %s", currentUser.ID, leagueID)
		return 0, common.ErrUnauthorized
	}

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

	// Get the player who owns this pokemon to check authorization
	ownerPlayer, err := s.playerRepo.GetPlayerByID(draftedPokemon.PlayerID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.ReleasePokemon) - Error getting owner player %s for drafted pokemon %s: %v", draftedPokemon.PlayerID, draftedPokemonID, err)
		return common.ErrInternalService
	}

	// Authorization: Admin, Commissioner of the league, or the player who owns the Pokemon.
	isCommissioner, err := s.isUserCommissioner(currentUser.ID, draftedPokemon.LeagueID)
	if err != nil {
		return err
	}

	if !currentUser.IsAdmin && !isCommissioner && currentUser.ID != ownerPlayer.UserID {
		log.Printf("(Error: DraftedPokemonService.ReleasePokemon) - Unauthorized attempt by user %s to release pokemon %s", currentUser.ID, draftedPokemonID)
		return common.ErrUnauthorized
	}

	if draftedPokemon.IsReleased {
		return errors.New("pokemon is already released")
	}

	err = s.draftedPokemonRepo.ReleasePokemon(draftedPokemonID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.ReleasePokemon) - Failed to release pokemon with ID %s: %v", draftedPokemonID, err)
		return common.ErrInternalService
	}

	return nil
}

// // re-drafts a released Pokemon (from free agents) to a new player.
// // this might be useless or just plain wrong
// func (s *draftedPokemonServiceImpl) ReDraftPokemon(currentUser *models.User, draftedPokemonID, newPlayerID uuid.UUID, newPickNumber int) error {
// 	// Authorization: Only Admin or Commissioner can redraft.
// 	isCommissioner, err := s.isUserCommissioner(currentUser.ID, draftedPokemonID)
// 	if err != nil {
// 		return err
// 	}
//
// 	if !currentUser.IsAdmin && !isCommissioner {
// 		log.Printf("(Error: DraftedPokemonService.ReDraftPokemon) - Unauthorized attempt by user %s to redraft pokemon %s", currentUser.ID, draftedPokemonID)
// 		return common.ErrUnauthorized
// 	}
//
// 	// Additional checks:
// 	// 1. Ensure the draftedPokemonID exists and is currently released.
// 	draftedPokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(draftedPokemonID)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return common.ErrDraftedPokemonNotFound
// 		}
// 		log.Printf("(Error: DraftedPokemonService.ReDraftPokemon) - Failed to get drafted pokemon %s: %v", draftedPokemonID, err)
// 		return fmt.Errorf("failed to get drafted pokemon: %w", err)
// 	}
// 	if !draftedPokemon.IsReleased {
// 		return errors.New("pokemon is not released and cannot be re-drafted")
// 	}
//
// 	// 2. Ensure newPlayerID exists and is a valid player in the same league.
// 	newPlayer, err := s.playerRepo.GetPlayerByID(newPlayerID)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return errors.New("new player not found")
// 		}
// 		log.Printf("(Error: ReDraftPokemon) - Failed to get new player %s: %v", newPlayerID, err)
// 		return fmt.Errorf("failed to get new player: %w", err)
// 	}
// 	if newPlayer.LeagueID != draftedPokemon.LeagueID {
// 		return errors.New("new player is not in the same league as the pokemon")
// 	}
//
// 	err = s.draftedPokemonRepo.ReDraftPokemon(draftedPokemonID, newPlayerID, newPickNumber)
// 	if err != nil {
// 		log.Printf("(Error: DraftedPokemonService.ReDraftPokemon) - Failed to re-draft pokemon with ID %s to player %s: %v", draftedPokemonID, newPlayerID, err)
// 		return fmt.Errorf("failed to re-draft pokemon: %w", err)
// 	}
//
// 	return nil
// }

// gets count of actively drafted Pokemon by a player.
func (s *draftedPokemonServiceImpl) GetDraftedPokemonCountByPlayer(currentUser *models.User, playerID uuid.UUID) (int64, error) {
	// First, get the target player to check their UserID and LeagueID.
	targetPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, common.ErrPlayerNotFound
		}
		log.Printf("(Error: GetDraftedPokemonCountByPlayer) - Failed to get target player %s: %v", playerID, err)
		return 0, common.ErrInternalService
	}

	// Authorization: Admin, or the player themselves, or a player in the same league.
	if !currentUser.IsAdmin {
		if currentUser.ID != targetPlayer.UserID { // Not admin and not viewing self
			isCurrentUserInTargetLeague, err := s.isUserPlayerInLeague(currentUser.ID, targetPlayer.LeagueID)
			if err != nil {
				return 0, err
			}
			if !isCurrentUserInTargetLeague {
				log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonCountByPlayer) - Unauthorized access to player %s's drafted pokemon count by user %s", playerID, currentUser.ID)
				return 0, common.ErrUnauthorized
			}
		}
	}

	count, err := s.draftedPokemonRepo.GetDraftedPokemonCountByPlayer(playerID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetDraftedPokemonCountByPlayer) - Failed to get drafted pokemon count for player %s: %v", playerID, err)
		return 0, common.ErrInternalService
	}

	return count, nil
}

// gets draft history for a league (all picks in order, including released).
func (s *draftedPokemonServiceImpl) GetDraftHistory(currentUser *models.User, leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	// Authorization check: Admin or a player in the league.
	isPlayerInLeague, err := s.isUserPlayerInLeague(currentUser.ID, leagueID)
	if err != nil {
		return nil, err
	}

	if !currentUser.IsAdmin && !isPlayerInLeague {
		log.Printf("(Error: DraftedPokemonService.GetDraftHistory) - Unauthorized attempt by user %s for league %s", currentUser.ID, leagueID)
		return nil, common.ErrUnauthorized
	}

	history, err := s.draftedPokemonRepo.GetDraftHistory(leagueID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.GetDraftHistory) - Failed to get draft history for league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}

	return history, nil
}

// TODO: this is very basic for now. what we want is a full blown trade offer system.
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

	// Authorization: Admin, Commissioner of the league, or the current owner of the Pokemon.
	isCommissioner, err := s.isUserCommissioner(currentUser.ID, draftedPokemon.LeagueID)
	if err != nil {
		return err
	}

	// Basic authorization: Admin, Commissioner, or the current owner can initiate/approve.
	// More complex trade logic (e.g., both players agree) would be implemented here or in a higher-level "Trade" service.
	if !currentUser.IsAdmin && !isCommissioner && currentUser.ID != currentOwnerPlayer.UserID {
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
	// Authorization: Admin, Commissioner of the league, or the player making their own draft pick.
	isCommissioner, err := s.isUserCommissioner(currentUser.ID, leagueID)
	if err != nil {
		return err
	}

	// Get the player entity for the current user in this specific league.
	// needed to verify if the current user is the player making the pick.
	currentPlayer, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User is not a player, cannot draft unless admin/commissioner
			if !currentUser.IsAdmin && !isCommissioner {
				return common.ErrUnauthorized
			}
		} else {
			log.Printf("(Error: DraftPokemonTransaction) - Failed to get player for user %s in league %s: %v", currentUser.ID, leagueID, err)
			return common.ErrInternalService
		}
	}

	// If not admin and not commissioner, ensure the user is drafting for themselves.
	if !currentUser.IsAdmin && !isCommissioner {
		if currentPlayer == nil || currentPlayer.ID != draftedPokemon.PlayerID {
			log.Printf("(Error: DraftPokemonTransaction) - Unauthorized attempt by user %s to draft for player %s in league %s (not self/admin/commissioner)", currentUser.ID, draftedPokemon.PlayerID, leagueID)
			return common.ErrUnauthorized
		}
	}
	// More specific draft-turn-based authorization logic would go here.
	// check if it's actually `currentPlayer`'s turn.

	// Add more specific validation/checks before starting the transaction, e.g.,
	// Does the league exist? (Implicitly checked by commissioner/player check)
	// Does the pokemon species exist in the league pool and is it available? (Can be checked here)
	// Does the player have enough draft points? (Can be checked here)

	err = s.draftedPokemonRepo.DraftPokemonTransaction(draftedPokemon, leagueID, pokemonSpeciesID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.DraftPokemonTransaction) - Failed to perform draft transaction for league %s, species %s: %v", leagueID, pokemonSpeciesID, err)
		return fmt.Errorf("failed to perform draft transaction: %w", err)
	}

	return nil
}

// soft deletes a drafted Pokemon entry.
func (s *draftedPokemonServiceImpl) DeleteDraftedPokemon(currentUser *models.User, draftedPokemonID uuid.UUID) error {
	// Get the drafted Pokemon to determine its league.
	draftedPokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(draftedPokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrDraftedPokemonNotFound
		}
		log.Printf("(Error: DeleteDraftedPokemon) - Error getting drafted pokemon %s for deletion: %v", draftedPokemonID, err)
		return common.ErrInternalService
	}

	// Authorization: Only Admin or Commissioner of the league can delete.
	isCommissioner, err := s.isUserCommissioner(currentUser.ID, draftedPokemon.LeagueID)
	if err != nil {
		return err
	}

	if !currentUser.IsAdmin && !isCommissioner {
		log.Printf("(Error: DraftedPokemonService.DeleteDraftedPokemon) - Unauthorized attempt by user %s to delete pokemon %s", currentUser.ID, draftedPokemonID)
		return common.ErrUnauthorized
	}

	err = s.draftedPokemonRepo.DeleteDraftedPokemon(draftedPokemonID)
	if err != nil {
		log.Printf("(Error: DraftedPokemonService.DeleteDraftedPokemon) - Failed to delete drafted pokemon with ID %s: %v", draftedPokemonID, err)
		return common.ErrInternalService
	}

	return nil
}
