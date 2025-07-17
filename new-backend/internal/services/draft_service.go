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
	MakePick(currentUser *models.User, league *models.League, leaguePokemonID uuid.UUID) error
	// SkipTurn(leagueID uuid.UUID, currentUser *models.User) (*models.Draft, error)
	// StartTradingPeriod(leagueID uuid.UUID, currentUser *models.User) (*models.League, error)
	// EndTradingPeriod(leagueID uuid.UUID, currentUser *models.User) (*models.League, error)
	// AddFreeAgencyPoints(leagueID uuid.UUID) error
	// DropFreeAgent(leagueID, draftedPokemonID uuid.UUID, currentUser *models.User) (*models.DraftedPokemon, error)
	// PickupFreeAgent(leagueID, pokemonSpeciesID uuid.UUID, currentUser *models.User) (*models.DraftedPokemon, error)
}

type draftServiceImpl struct {
	draftRepo          repositories.DraftRepository
	leagueRepo         repositories.LeagueRepository
	playerRepo         repositories.PlayerRepository
	leaguePokemonRepo  repositories.LeaguePokemonRepository
	draftedPokemonRepo repositories.DraftedPokemonRepository
	webhookService     *WebhookService
}

func NewDraftService(
	leagueRepo repositories.LeagueRepository,
	leaguePokemonRepo repositories.LeaguePokemonRepository,
	draftRepo repositories.DraftRepository,
	draftedPokemonRepo repositories.DraftedPokemonRepository,
	playerRepo repositories.PlayerRepository,
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
	league, err := (s.leagueRepo).GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("(Error: DraftService.StartDraft) - Could not get league %s: %v\n", leagueID, err)
		return nil, common.ErrLeagueNotFound
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

// assuming the controller fetches the draft to send it here
// only used for the initial draft (not free agent transactions)
func (s *draftServiceImpl) MakePick(currentUser *models.User, league *models.League, leaguePokemonID uuid.UUID) error {
	// fetch draft from league
	draft, err := s.draftRepo.GetDraftByLeagueID(league.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("DraftService: MakePick - draft for leagueID %s not found: %w\n", league.ID, err)
			return common.ErrDraftNotFound
		}
		log.Printf("DraftService: MakePick - Could not fetch draft: %w\n", err)
		return common.ErrInternalService
	}

	// check league state
	if status := league.Status; status.IsValid() && status != models.LeagueStatusDrafting {
		log.Printf("DraftService: MakePick - user %s tried to draft when league %s not in drafting status: %v\n", currentUser.ID, league.ID, err)
		return common.ErrInvalidState
	}

	if status := draft.Status; status.IsValid() && status != models.DraftStatusStarted {
		log.Printf("DraftService: MakePick - user %s tried to draft when league %s not in drafting status (%s): %v\n", currentUser.ID, league.ID, status, err)
		return common.ErrInvalidState

	}

	player, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, league.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("DraftService: MakePick - Player of userID %s in league %s not found: %w\n", currentUser.ID, league.ID, err)
			return common.ErrPlayerNotFound
		}
		return common.ErrInternalService
	}

	// check if it's the right player's turn
	if currentTurnPlayerID := *draft.CurrentTurnPlayerID; currentTurnPlayerID != player.ID {
		log.Printf("DraftService: MakePick - player %s tried to draft when it isn't their turn. Current Turn: Player %s\n", currentTurnPlayerID, *draft.CurrentTurnPlayerID)
		return common.ErrUnauthorized
	}

	// check if the pokemon picked is valid (has to check with all the different models)
	leaguePokemon, err := s.leaguePokemonRepo.GetLeaguePokemonByID(leaguePokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("DraftService: MakePick - no corresponding league pokemon %s found in league %s : %w\n", leaguePokemonID, league.ID, err)
			return common.ErrLeaguePokemonNotFound
		}
		return common.ErrInternalService
	}

	// if the mon isn't availble
	if !leaguePokemon.IsAvailable {
		log.Printf("DraftService: MakePick - League Pokemon %s is not availble\n", leaguePokemon.ID)
		return common.ErrConflict
	}

	// check if player has enough DraftPoints
	if leaguePokemon.Cost == nil || player.DraftPoints < *leaguePokemon.Cost { // if they do not
		return common.ErrInsufficientDraftPoints
	}

	// get next overall draft pick number to assign to the DraftedPokemon record (1-based)
	nextOverallPickNumber, err := s.draftedPokemonRepo.GetNextDraftPickNumber(leaguePokemonID)
	if err != nil {
		log.Printf("DraftService: MakePick - failed to get next overall draft pick number for league %s.\n", league.ID)
		return common.ErrInternalService
	}

	playerCount, err := s.playerRepo.GetPlayerCountByLeague(league.ID)
	if err != nil {
		log.Printf("DraftService: MakePick - failed to get player count for league %s.\n", league.ID)
		return common.ErrInternalService
	}
	if playerCount == 0 { // this should never happen if the draft has started but gotta make sure
		log.Printf("DraftService: MakePick - no players in league %d. (Unreachable Control Flow)\n", playerCount)
		return common.ErrInternalService
	}

	// set the round number
	draftRoundNumber := ((nextOverallPickNumber - 1) / int(playerCount)) + 1

	// now create the actual draftedPokemon
	draftedPokemon := &models.DraftedPokemon{
		LeagueID:         league.ID,
		PlayerID:         player.ID,
		PokemonSpeciesID: leaguePokemon.PokemonSpeciesID,
		DraftRoundNumber: draftRoundNumber,
		DraftPickNumber:  nextOverallPickNumber,
		IsReleased:       false,
	}

	// store current player's draft points in case of failure (gorm should revert changes but we're making sure)
	playerDraftPointsBeforeTransaction := player.DraftPoints

	// perform the transaction (should adjust player's points if succeeds and revert if it doesn't)
	err = s.draftedPokemonRepo.DraftPokemonTransaction(draftedPokemon)
	if err != nil {
		log.Printf("DraftService: MakePick - Player %s (League %s) draft transaction for LeaguePokemon %s failed: %w\n", player.ID, league.ID, leaguePokemon.ID, err)
		if err2 := s.playerRepo.UpdatePlayerDraftPoints(player.ID, playerDraftPointsBeforeTransaction); err2 != nil {
			log.Printf("DraftService: MakePick - Fallback Draft Points Set also failed for player %s (league %s): %w\n", player.ID, league.ID, leaguePokemon.ID, err)
			log.Printf("If this was reached the whole fucking backend deserves to be nuked because holy shit. NUKE FUCKING EVERYTHING. THE BACKEND, THE DATABASE. EVERYTHING. START AGAIN.\n")
			log.Printf("THE BACKEND SERVER DESERVES TO FUCKING DIE.\n")
			log.Printf("WHY STOP THERE THO? you might as well sudo rm -rf --no-preserve-root\n")
		}
		return common.ErrInternalService
	}

	// update draft model
	// get all players to change set the CurrentPlayer's turn for the next one
	allPlayers, err := s.playerRepo.GetPlayersByLeague(draft.LeagueID)
	if err != nil {
		log.Printf("DraftService: MakePick - Could not get all players in league %s: %w\n", league.ID, err)
		return common.ErrInternalService
	}

	draft.CurrentPickInRound++
	currentPlayerIdx := -1
	for i, p := range allPlayers {
		if p.ID == player.ID {
			currentPlayerIdx = i
			break
		}
	}
	// there's no reason for playeridx to be -1 if it was found before so i refuse to do it

	var nextPlayerIdx int
	if league.IsSnakeRoundDraft {
		// ternaries would be so cool in this langauge
		if draft.CurrentRound%2 == 0 { // even rounds are reverse order
			nextPlayerIdx = currentPlayerIdx - 1
		} else {
			nextPlayerIdx = currentPlayerIdx + 1
		}
	} else {
		nextPlayerIdx = currentPlayerIdx + 1
	}

	// check if the round is over
	if nextPlayerIdx >= int(playerCount) || nextPlayerIdx < 0 { // Changed > to >= to handle 0-based index correctly
		draft.CurrentRound++
		draft.CurrentPickInRound = 1                               // reset pick order
		if league.IsSnakeRoundDraft && draft.CurrentRound%2 == 0 { // if snake round drafting and an even round
			nextPlayerIdx = int(playerCount) - 1 // it's an index hence - 1 despite current pick being 0 based
		} else {
			nextPlayerIdx = 0
		}
		// TODO: check for draft completion (if max rounds has completed)
		// ensure status is changed etc.
	}

	// finally set the next turn of player
	nextTurnPlayer := allPlayers[nextPlayerIdx]
	draft.CurrentTurnPlayerID = &nextTurnPlayer.ID
	draft.CurrentTurnStartTime = func() *time.Time { t := time.Now(); return &t }() // done with a lambda because it  expects a pointer due to it being null

	if err := s.draftRepo.UpdateDraft(draft); err != nil {
		log.Printf("DraftService: MakePick - Failed to update draft: %w\n", err)
		// TODO: Some sort of compensation needs to be here/fixing of the state needs to happen here
		return fmt.Errorf("Failed to update draft state: %w", err)
	}

	// TODO: Trigger webhook notification for the pick that just happened as well as the turn change

	return nil // no errors

	// if future me is reading this and has read it the whole way through AND understood what's going on here.
	// you should give yourself more credit
	// this needs to be broken down to several smaller functions. Maybe YOU should do that :)))))).
}
