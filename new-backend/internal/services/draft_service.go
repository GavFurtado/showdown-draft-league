package services

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
)

type DraftService interface {
	StartDraft(currentUser *models.User, leagueID uuid.UUID) (*models.Draft, error)
	// MakePick(leagueID, pokemonSpeciesID uuid.UUID, currentUser *models.User) (*models.Draft, error)
	// SkipTurn(leagueID uuid.UUID, currentUser *models.User) (*models.Draft, error)
	// StartTradingPeriod(leagueID uuid.UUID, currentUser *models.User) (*models.League, error)
	// EndTradingPeriod(leagueID uuid.UUID, currentUser *models.User) (*models.League, error)
	// AddFreeAgencyPoints(leagueID uuid.UUID) error
	// DropFreeAgent(leagueID, draftedPokemonID uuid.UUID, currentUser *models.User) (*models.DraftedPokemon, error)
	// PickupFreeAgent(leagueID, pokemonSpeciesID uuid.UUID, currentUser *models.User) (*models.DraftedPokemon, error)
}

type draftServiceImpl struct {
	draftRepo          repositories.DraftRepository // turns out when implemented with interfaces you dont do pointer
	leagueRepo         *repositories.LeagueRepository
	playerRepo         *repositories.PlayerRepository
	leaguePokemonRepo  *repositories.LeaguePokemonRepository
	draftedPokemonRepo *repositories.DraftedPokemonRepository
	webhookService     *WebhookService
}

func NewDraftService(
	leagueRepo *repositories.LeagueRepository,
	leaguePokemonRepo *repositories.LeaguePokemonRepository,
	draftRepo repositories.DraftRepository,
	draftedPokemonRepo *repositories.DraftedPokemonRepository,
	playerRepo *repositories.PlayerRepository,
	webhookService *WebhookService,
) DraftService {
	return &draftServiceImpl{
		draftRepo:          draftRepo,
		leagueRepo:         leagueRepo,
		playerRepo:         playerRepo,
		leaguePokemonRepo:  leaguePokemonRepo,
		draftedPokemonRepo: draftedPokemonRepo,
		webhookService:     webhookService,
	}
}

// --- Private Helper Authorization Methods ---
// These encapsulate error handling for repository calls
func (s *draftServiceImpl) isUserCommissioner(userID, leagueID uuid.UUID) (bool, error) {
	isComm, err := s.leagueRepo.IsUserCommissioner(userID, leagueID)
	if err != nil {
		log.Printf("(Error: draftedPokemonService.isUserCommissioner) - Failed to check commissioner status for user %s in league %s: %v", userID, leagueID, err)
		return false, fmt.Errorf("failed to check commissioner status: %w", err)
	}
	return isComm, nil
}

func (s *draftServiceImpl) isUserPlayerInLeague(userID, leagueID uuid.UUID) (bool, error) {
	isPlayer, err := s.leagueRepo.IsUserPlayerInLeague(userID, leagueID)
	if err != nil {
		log.Printf("(Error: draftedPokemonService.isUserPlayerInLeague) - Failed to check player status for user %s in league %s: %v", userID, leagueID, err)
		return false, fmt.Errorf("failed to check player status: %w", err)
	}
	return isPlayer, nil
}

func (s *draftServiceImpl) StartDraft(currentUser *models.User, leagueID uuid.UUID) (*models.Draft, error) {
	// Retrieve the league
	league, err := (*s.leagueRepo).GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("(Error: DraftService.StartDraft) - Could not get league %s: %v\n", leagueID, err)
		return nil, common.ErrInternalService
	}

	if league == nil {
		return nil, common.ErrLeagueNotFound
	}

	// Check if currentUser is the commissioner
	if league.CommissionerUserID != currentUser.ID {
		log.Printf("(Error: DraftService.StartDraft) - User %s is not the commissioner of league %s\n", currentUser.ID, leagueID)
		return nil, common.ErrUnauthorized
	}

	// Retrieve players in the league, sorted by draft position
	players, err := s.playerRepo.GetPlayersByLeague(leagueID)
	if err != nil {
		log.Printf("(Error: DraftService.StartDraft) - Could not get players for league %s: %v\n", leagueID, err)
		return nil, common.ErrInternalService
	}

	if len(players) == 0 {
		log.Printf("(Error: DraftService.StartDraft) - No players found for league %s\n", leagueID)
		return nil, errors.New("cannot start draft with no players") // Consider a more specific error
	}

	// Initialize the Draft model
	firstPlayerID := players[0].ID
	currTime := time.Now()

	draft := &models.Draft{
		LeagueID:                    leagueID,
		Status:                      models.DraftStatusStarted,
		CurrentRound:                1,
		CurrentPickInRound:          1,
		CurrentTurnPlayerID:         &firstPlayerID,
		CurrentTurnStartTime:        &currTime,
		PlayersWithAccumulatedPicks: make(map[uuid.UUID]int),
		// TurnTimeLimit and IsSnakeRoundDraft will be inherited or set based on league rules, maybe fetch League to get these?
	}

	// Save the Draft model
	if err := s.draftRepo.CreateDraft(draft); err != nil {
		log.Printf("(Error: DraftService.StartDraft) - Failed to create draft for league %s: %v\n", leagueID, err)
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}

	// Update the league status to DRAFTING
	league.Status = models.LeagueStatusDrafting
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("(Error: DraftService.StartDraft) - Failed to update league status for league %s: %v\n", leagueID, err)
		// TODO: Consider rolling back draft creation if this fails
		return nil, fmt.Errorf("failed to update league status: %w", err)
	}

	// Send an initial webhook notification
	// TODO: Implement webhook message creation logic
	// if err := (*s.webhookService).SendWebhookMessage(league.DiscordWebhookURL, "Draft has started!"); err != nil {
	// 	log.Printf("(Warning: DraftService.StartDraft) - Failed to send webhook for league %s: %v\n", leagueID, err)
	// 	// Continue execution, webhook failure shouldn't stop the draft
	// }

	return draft, nil
}

/**
2.  **Refactor `DraftService` Core:**
    *   **Implement Main Draft Flow:** The `MakePick` function will be enhanced to handle the full pick process:
        *   Fetching the current `Draft` and associated `League` state.
        *   Validating the draft status (must be `STARTED`).
        *   Authorizing the user (must be the current turn player or commissioner/admin).
        *   Validating the selected `PokemonSpecies`: ensure it exists, is in the league's available pool (`LeaguePokemon`), and is not already drafted (unless it's a free agent pickup).
        *   Getting the next sequential draft pick number for the league.
        *   Creating the `models.DraftedPokemon` instance for the pick.
        *   Executing the database operations atomically (transaction) to:
            *   Create the `DraftedPokemon` record.
            *   Mark the corresponding `LeaguePokemon` entry as unavailable (`IsAvailable = false`).
            *   Update the `Draft` model (increment pick number, update round, set next turn player, update turn start time).
            *   Update the `Player`'s draft points if applicable (e.g., for free agency pickups).
        *   Handling the transition to the next turn, including calculating the next player based on snake/non-snake format.
        *   Checking for draft completion and updating statuses if necessary.
        *   Triggering webhook notifications for successful picks and turn changes.
*/

// assuming the controller fetches the draft to send it here
func (s *draftServiceImpl) MakePick(currentUser *models.User, league *models.League, pokemonSpeciesID uuid.UUID) error {
	// fetch draft from league
	draft, err := s.draftRepo.GetDraftByLeagueID(league.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("DraftService: MakePick - draft for leagueID %d not found: %w\n", league.ID, err)
			return common.ErrDraftNotFound
		}
		log.Printf("DraftService: MakePick - Could not fetch draft: %w\n", err)
		return common.ErrInternalService
	}

	// check league state
	if status := league.Status; status.IsValid() && status != models.LeagueStatusDrafting {
		log.Printf("DraftService: MakePick - user %d tried to draft when league %d not in drafting status: %v\n", currentUser.ID, league.ID, err)
		return common.ErrInvalidState
	}

	if status := draft.Status; status.IsValid() && status != models.DraftStatusStarted {
		log.Printf("DraftService: MakePick - user %d tried to draft when league %d not in drafting status (%s): %v\n", currentUser.ID, league.ID, status, err)
		return common.ErrInvalidState

	}

	player, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, league.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("DraftService: MakePick - Player of userID %d in league %d not found: %w\n", currentUser.ID, league.ID, err)
			return common.ErrPlayerNotFound
		}
		return common.ErrInternalService
	}

	// check if it's the right player's turn
	if currentTurnPlayerID := *draft.CurrentTurnPlayerID; currentTurnPlayerID != player.ID {
		log.Printf("DraftService: MakePick - player %d tried to draft when it isn't their turn. Current Turn: Player %d\n", currentTurnPlayerID)
		return common.ErrUnauthorized
	}

	// check if the pokemon picked is valid (has to check with all the different models)
	leaguePokemon, err := s.leaguePokemonRepo.GetLeaguePokemonBySpecies(league.ID, pokemonSpeciesID)
	if err != nil {

	}

}
