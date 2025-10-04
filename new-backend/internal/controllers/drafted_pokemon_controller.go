package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
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
}

type draftedPokemonControllerImpl struct {
	draftedPokemonService services.DraftedPokemonService
}

func NewDraftedPokemonController(draftedPokemonService services.DraftedPokemonService) *draftedPokemonControllerImpl {
	return &draftedPokemonControllerImpl{
		draftedPokemonService: draftedPokemonService,
	}
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
func (c *draftedPokemonControllerImpl) GetReleasedDraftedPokemonByLeaguj(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, isPokemonDrafted)
}
