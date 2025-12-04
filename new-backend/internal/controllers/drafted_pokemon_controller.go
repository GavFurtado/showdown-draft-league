package controllers

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strconv"
)

type DraftedPokemonController interface {
	// GET single drafted pokemon by its ID
	GetDraftedPokemonByID(ctx *gin.Context)
	// GET all drafted pokemon by player with :playerId (includes released)
	GetDraftedPokemonByPlayer(ctx *gin.Context)
	// GET all drafted pokemon in a league (includes released)
	GetDraftedPokemonByLeague(ctx *gin.Context)
	// GET all drafted pokemon in a league (excludes released)
	GetActiveDraftedPokemonByLeague(ctx *gin.Context)
	// GET all RELEASED drafted pokemon in a league
	GetReleasedPokemonByLeague(ctx *gin.Context)
	// GET if the species :speciesId has been drafted for this league :leagueId
	IsPokemonDrafted(ctx *gin.Context)
	// GET next draft pick number for the league :leagueId (if league.Status == "DRAFTING")
	GetNextDraftPickNumber(ctx *gin.Context)
	// PATCH Marks a pokemon as Released
	ReleasePokemon(ctx *gin.Context)
	// GET the number of active draftedPokmon has made (how many pokemon on player's roster)
	GetDraftedPokemonCountByPlayer(ctx *gin.Context)
	// GET draft history for a league (all picks in order, including released and includes transfers).
	GetDraftHistory(ctx *gin.Context)
	DropPokemon(ctx *gin.Context)
	PickupFreeAgent(ctx *gin.Context)
}

type draftedPokemonControllerImpl struct {
	draftedPokemonService services.DraftedPokemonService
}

func NewDraftedPokemonController(draftedPokemonService services.DraftedPokemonService) DraftedPokemonController {
	return &draftedPokemonControllerImpl{
		draftedPokemonService: draftedPokemonService,
	}
}

func (c *draftedPokemonControllerImpl) DropPokemon(ctx *gin.Context) {
	currentUser, err := c.getUserFromContext(ctx)
	if err != nil {
		return // response already sent
	}

	draftedPokemonID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	if err := c.draftedPokemonService.DropPokemon(currentUser, draftedPokemonID); err != nil {
		log.Printf("LOG: (DraftedPokemonController: DropPokemon) - Service method error: %v\n", err)
		switch err {
		case common.ErrDraftedPokemonNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case common.ErrInvalidState:
			ctx.JSON(http.StatusConflict, gin.H{"error": "League is not in a transfer window"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "pokemon dropped successfully"})
}

func (c *draftedPokemonControllerImpl) PickupFreeAgent(ctx *gin.Context) {
	currentUser, err := c.getUserFromContext(ctx)
	if err != nil {
		return // response already sent
	}

	leaguePokemonID, err := uuid.Parse(ctx.Param("leaguePokemonId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	if err := c.draftedPokemonService.PickupFreeAgent(currentUser, leaguePokemonID); err != nil {
		log.Printf("LOG: (DraftedPokemonController: PickupFreeAgent) - Service method error: %v\n", err)
		switch err {
		case common.ErrLeaguePokemonNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrInsufficientTransferCredits:
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case common.ErrInvalidState:
			ctx.JSON(http.StatusConflict, gin.H{"error": "League is not in a transfer window"})
		case common.ErrConflict:
			ctx.JSON(http.StatusConflict, gin.H{"error": "Pokemon is not available to sign"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "free agent signed successfully"})
}

// Helpers
func (c *draftedPokemonControllerImpl) getUserFromContext(ctx *gin.Context) (*models.User, error) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("LOG: (DraftedPokemonController: getUserFromContext) - no user in context\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrNoUserInContext.Error()})
		return nil, common.ErrNoUserInContext
	}
	return currentUser, nil
}

// GET api/leagues/:leagueId/drafted_pokemon/:id/
// GET single drafted pokemon by its ID
// player permission: rbac.PermissionReadDraftedPokemon
func (c *draftedPokemonControllerImpl) GetDraftedPokemonByID(ctx *gin.Context) {
	pokemonID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	draftedPokemon, err := c.draftedPokemonService.GetDraftedPokemonByID(pokemonID)
	if err != nil {
		log.Printf("LOG: (PlayerController: GetDraftedPokemonByPlayer) - Service method error: %v\n", err)
		switch err {
		case common.ErrDraftedPokemonNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrDraftedPokemonNotFound.Error()})
		default:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, draftedPokemon)
}

// GET api/leagues/:leagueId/drafted_pokemon/player/:playerId/
// GET all pokemon drafted by the player :playerId
// player permission: rbac.PermissionReadDraftedPokemon
func (c *draftedPokemonControllerImpl) GetDraftedPokemonByPlayer(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	draftedPokemonList, err := c.draftedPokemonService.GetDraftedPokemonByPlayer(playerID)
	if err != nil {
		log.Printf("LOG: (DraftedPokemonController: GetDraftedPokemonByPlayer) - Service method error: %v\n", err)
		switch err {
		case common.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrPlayerNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, draftedPokemonList)
}

// GET api/leagues/:leagueId/drafted_pokemon/
// GET all pokemon drafted in a league
// player permission: rbac.PermissionReadDraftedPokemon
func (c *draftedPokemonControllerImpl) GetDraftedPokemonByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	allDraftedPokemon, err := c.draftedPokemonService.GetDraftedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("LOG: (DraftedPokemonController: GetDraftedPokemonByLeague) - Service method error: %v\n", err)
		switch err {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, allDraftedPokemon)
}

// GET api/leagues/:leagueId/drafted_pokemon/active
// GET all pokemon drafted in a league (excludes released)
// player permission: rbac.PermissionReadDraftedPokemon
func (c *draftedPokemonControllerImpl) GetActiveDraftedPokemonByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	allActiveDraftedPokemon, err := c.draftedPokemonService.GetActiveDraftedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("LOG: (DraftedPokemonController: GetActiveDraftedPokemonByLeague) - Service method error: %v\n", err)
		switch err {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, allActiveDraftedPokemon)
}

// GET api/leagues/:leagueId/drafted_pokemon/released
// GET all RELEASED pokemon drafted in a league
// player permission: rbac.PermissionReadDraftedPokemon
func (c *draftedPokemonControllerImpl) GetReleasedPokemonByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	allReleasedDraftedPokemon, err := c.draftedPokemonService.GetReleasedPokemonByLeague(leagueID)
	if err != nil {
		log.Printf("LOG: (DraftedPokemonController: GetReleasedDraftedPokemonByLeague) - Service method error: %v\n", err)
		switch err {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, allReleasedDraftedPokemon)
}

// GET api/leagues/:leagueId/drafted_pokemon/is_drafted/:speciesId
// GET if the species :speciesId has been drafted for this league :leagueId
// player permission: rbac.PermissionReadDraftedPokemon
func (c *draftedPokemonControllerImpl) IsPokemonDrafted(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	speciesID, err2 := strconv.ParseInt(ctx.Param("speciesId"), 10, 64)
	if err != nil || err2 != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	isPokemonDrafted, err := c.draftedPokemonService.IsPokemonDrafted(leagueID, speciesID)
	if err != nil {
		log.Printf("LOG: (DraftedPokemonController: IsPokemonDrafted) - Service method error: %v\n", err)
		switch err {
		case common.ErrPokemonSpeciesNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrPokemonSpeciesNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"is_pokemon_drafted": isPokemonDrafted})
}

// GET api/leagues/:leagueId/draft/next_pick_number
// GET next draft pick number for the league :leagueID (if league.Status == "DRAFTING")
// player permission: rbac.PermissionReadDraft
// (yes this is misplaced; should be part of the Draft not DraftedPokemon)
func (c *draftedPokemonControllerImpl) GetNextDraftPickNumber(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	// league status checked in service method
	nextPickNumber, err := c.draftedPokemonService.GetNextDraftPickNumber(leagueID)
	if err != nil {
		log.Printf("LOG: (DraftedPokemonController: GetNextDraftPickNumber) - Service method error: %v\n", err)
		switch err {
		case common.ErrInvalidState:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "league status is not drafting"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"next_pick_number": nextPickNumber})
}

// PATCH api/leagues/:leagueId/drafted_pokemon/:id/release
// Marks a drafted pokemon as released
// player permission: rbac.PermissionCreateDraftedPokemon
func (c *draftedPokemonControllerImpl) ReleasePokemon(ctx *gin.Context) {
	currentUser, err := c.getUserFromContext(ctx)
	if err != nil {
		return // response already sent in this case
	}

	draftedPokemonID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	if err := c.draftedPokemonService.ReleasePokemon(currentUser, draftedPokemonID); err != nil {
		log.Printf("LOG: (DraftedPokemonController: ReleasePokemon) - Service method error: %v\n", err)
		switch err {
		case common.ErrDraftedPokemonNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrDraftedPokemonNotFound.Error()})
		case common.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrPlayerNotFound.Error()})
		case common.ErrPokemonAlreadyReleased:
			ctx.JSON(http.StatusConflict, gin.H{"error": common.ErrDraftedPokemonNotFound.Error()})
		case common.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrUnauthorized.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "operation success"})
}

// GET api/leagues/:leagueId/drafted_pokemon/count/:playerId
// GET the number of active draftedPokmon has made (how many pokemon on player's roster)
// player permission: rbac.PermissionReadDraftedPokemon
func (c *draftedPokemonControllerImpl) GetDraftedPokemonCountByPlayer(ctx *gin.Context) {
	currentUser, err := c.getUserFromContext(ctx)
	if err != nil {
		return
	}

	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	count, err := c.draftedPokemonService.GetDraftedPokemonCountByPlayer(currentUser, playerID)
	if err != nil {
		log.Printf("LOG: (DraftedPokemonController: GetDraftedPokemonCountByPlayer) - Service method error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// GET api/leagues/:leagueId/draft/history
// GET draft history for a league (all picks in order, including released and includes transfers).
// player permission: rbac.PermissionReadDraft
func (c *draftedPokemonControllerImpl) GetDraftHistory(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	draftHistory, err := c.draftedPokemonService.GetDraftHistory(leagueID)
	if err != nil {
		log.Printf("LOG: (DraftedPokemonController: GetDraftedPokemonCountByPlayer) - Service method error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		return
	}

	ctx.JSON(http.StatusOK, draftHistory)
}
