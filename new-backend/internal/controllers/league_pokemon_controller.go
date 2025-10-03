package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LeaguePokemonController handles CRUD HTTP requests on LeaguePokemon
type LeaguePokemonController interface {
	// POST to create a new pokemon for a league
	CreatePokemonForLeague(ctx *gin.Context)
	// POST to create many new pokemon for a league
	BatchCreatePokemonForLeague(ctx *gin.Context)
	// PUT to update LeaguePokemon for a league
	UpdateLeaguePokemon(ctx *gin.Context)
	// GETs all LeaguePokemon for a league
	GetAllPokemonByLeague(ctx *gin.Context)

	// more to be implemented
}

type leaguePokemonControllerImpl struct {
	leaguePokemonService services.LeaguePokemonService
}

func NewLeaguePokemonSpeciesController(leaguePokemonService services.LeaguePokemonService) *leaguePokemonControllerImpl {
	return &leaguePokemonControllerImpl{
		leaguePokemonService: leaguePokemonService,
	}
}

// POST api/leagues/:leagueID/pokemon/single/
func (c *leaguePokemonControllerImpl) CreatePokemonForLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("LOG: (Controller: CreatePokemonForLeague) - error: no user in context\n")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrNoUserInContext.Error()})
		return
	}

	var req common.LeaguePokemonCreateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	fmt.Printf("INFO: (Controller: CreatePokemonForLeague) - Received league pokemon creation request: %v\n", req)
	leaguePokemon, err := c.leaguePokemonService.CreatePokemonForLeague(currentUser, &req)
	if err != nil {
		log.Printf("LOG: (Controller: CreatePokemonForLeague) - Service method Error: %v\n", err)
		switch err {
		case common.ErrInvalidState:
			ctx.JSON(http.StatusForbidden, gin.H{"error": common.ErrInvalidState.Error()})
		case common.ErrLeagueNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrLeagueNotFound.Error()})
		case common.ErrPokemonSpeciesNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrPokemonSpeciesNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	fmt.Printf("INFO: (Controller: CreatePokemonForLeague) - Created LeaguePokemon %s(%s) successfully for league %s.\n",
		leaguePokemon.ID, leaguePokemon.PokemonSpecies.Name, leaguePokemon.LeagueID)

	ctx.JSON(http.StatusOK, leaguePokemon)
}

func (c *leaguePokemonControllerImpl) BatchCreatePokemonForLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("LOG: (Controller: BatchCreatePokemonForLeague) - error: no user in context\n")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrNoUserInContext.Error()})
		return
	}

	var req []common.LeaguePokemonCreateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	fmt.Printf("INFO: (Controller: BatchCreatePokemonForLeague) - Received batch league pokemon creation request with %d items\n", len(req))
	leaguePokemon, err := c.leaguePokemonService.BatchCreatePokemonForLeague(currentUser, req)
	if err != nil {
		log.Printf("LOG: (Controller: BatchCreatePokemonForLeague) - Service method Error: %v\n", err)
		switch err {
		case common.ErrInvalidState:
			ctx.JSON(http.StatusForbidden, gin.H{"error": common.ErrInvalidState.Error()})
		case common.ErrLeagueNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrLeagueNotFound.Error()})
		case common.ErrPokemonSpeciesNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrPokemonSpeciesNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	fmt.Printf("INFO: (Controller: BatchCreatePokemonForLeague) - Created %d LeaguePokemon successfully.\n",
		len(leaguePokemon))

	ctx.JSON(http.StatusOK, leaguePokemon)
}

func (c *leaguePokemonControllerImpl) UpdateLeaguePokemon(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("LOG: (Controller: BatchCreatePokemonForLeague) - error: no user in context\n")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrNoUserInContext.Error()})
		return
	}

	var req common.LeaguePokemonUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	}

	fmt.Printf("INFO: (Controller: UpdateLeaguePokemon) - Received update league pokemon request: %v\n", req)
	updatedLeaguePokemon, err := c.leaguePokemonService.UpdateLeaguePokemon(currentUser, &req)
	if err != nil {
		log.Printf("LOG: (Controller: UpdateLeaguePokemon) - Service method Error: %v\n", err)
		switch err {
		case common.ErrLeaguePokemonNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrLeaguePokemonNotFound.Error()})
		case common.ErrLeagueNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrLeagueNotFound.Error()})
		case common.ErrPokemonSpeciesNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrPokemonSpeciesNotFound.Error()})
		case common.ErrInvalidState:
			ctx.JSON(http.StatusForbidden, gin.H{"error": common.ErrInvalidState.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	fmt.Printf("INFO: (LeaguePokemonController: BatchCreatePokemonForLeague) - Updated LeaguePokemon %s(%s) successfully.\n",
		updatedLeaguePokemon.ID, updatedLeaguePokemon.PokemonSpecies.Name)
	ctx.JSON(http.StatusOK, updatedLeaguePokemon)
}

// GET api/leagues/:leagueId/all
// GETs all LeaguePokemon for a league
// Player permission required: rbac.PermissionReadLeaguePokemon
func (c *leaguePokemonControllerImpl) GetAllPokemonByLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("LOG: (LeaguePokemonController: BatchCreatePokemonForLeague) - error: no user in context\n")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrNoUserInContext.Error()})
		return
	}

	leagueIDstr := ctx.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDstr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	fmt.Printf("INFO: (Controller: UpdateLeaguePokemon) - Received valid GetAllPokemonByLeague request (user %s).\n", currentUser.ID)
	leaguePokemon, err := c.leaguePokemonService.GetAllPokemonInLeague(currentUser, leagueID)
	if err != nil {
		switch err {
		case common.ErrLeagueNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrLeagueNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, leaguePokemon)
}
