package controllers

import (
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PoolEntryController interface {
	GetByID(ctx *gin.Context)
	GetByLeague(ctx *gin.Context)
	GetAvailableByLeague(ctx *gin.Context)
	Create(ctx *gin.Context)
	CreateBatch(ctx *gin.Context)
	Update(ctx *gin.Context)
}

type poolEntryControllerImpl struct {
	poolEntryService services.PoolEntryService
}

func NewPoolEntryController(poolEntryService services.PoolEntryService) PoolEntryController {
	return &poolEntryControllerImpl{
		poolEntryService: poolEntryService,
	}
}

// GetByID godoc
//
//	@Summary		Get a pool entry
//	@Description	Pool entry details
//	@Tags			Pool Entries
//	@Produce		json
//	@Param			id	path		string	true	"Pool Entry ID"
//	@Success		200	{object}	models.PoolEntry
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/pool-entries/{id} [get]
func (c *poolEntryControllerImpl) GetByID(ctx *gin.Context) {
	entryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	entry, err := c.poolEntryService.GetByID(entryID)
	if err != nil {
		log.Printf("LOG: (PoolEntryController: GetByID) - Service method error: %v\n", err)
		switch err {
		case types.ErrPoolEntryNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrPoolEntryNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

// GetByLeague godoc
//
//	@Summary		Get pool entries by league
//	@Description	All pool entries for a league
//	@Tags			Pool Entries
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.PoolEntry
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/pool-entries [get]
func (c *poolEntryControllerImpl) GetByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	entries, err := c.poolEntryService.GetByLeague(leagueID)
	if err != nil {
		log.Printf("LOG: (PoolEntryController: GetByLeague) - Service method error: %v\n", err)
		switch err {
		case types.ErrLeagueNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrLeagueNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, entries)
}

// GetAvailableByLeague godoc
//
//	@Summary		Get available pool entries
//	@Description	Unclaimed Pokemon available in the pool
//	@Tags			Pool Entries
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.PoolEntry
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/pool-entries/available [get]
func (c *poolEntryControllerImpl) GetAvailableByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	entries, err := c.poolEntryService.GetAvailableByLeague(leagueID)
	if err != nil {
		log.Printf("LOG: (PoolEntryController: GetAvailableByLeague) - Service method error: %v\n", err)
		switch err {
		case types.ErrLeagueNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrLeagueNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, entries)
}

// Create godoc
//
//	@Summary		Create a pool entry
//	@Description	Add a Pokemon to the draft pool
//	@Tags			Pool Entries
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string								true	"League ID"
//	@Param			request		body		requests.PoolEntryCreateRequestDTO	true	"Pool entry"
//	@Success		200			{object}	models.PoolEntry
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		403			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/pool-entries/single [post]
func (c *poolEntryControllerImpl) Create(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrNoUserInContext.Error()})
		return
	}

	var req requests.PoolEntryCreateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	entry, err := c.poolEntryService.Create(currentUser, &req)
	if err != nil {
		log.Printf("LOG: (PoolEntryController: Create) - Service method error: %v\n", err)
		switch err {
		case types.ErrInvalidState:
			ctx.JSON(http.StatusForbidden, gin.H{"error": types.ErrInvalidState.Error()})
		case types.ErrLeagueNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrLeagueNotFound.Error()})
		case types.ErrPokemonSpeciesNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrPokemonSpeciesNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

// CreateBatch godoc
//
//	@Summary		Create pool entries in batch
//	@Description	Add multiple Pokemon to the draft pool
//	@Tags			Pool Entries
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string									true	"League ID"
//	@Param			request		body		[]requests.PoolEntryCreateRequestDTO	true	"Pool entries"
//	@Success		200			{array}		models.PoolEntry
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		403			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/pool-entries/batch [post]
func (c *poolEntryControllerImpl) CreateBatch(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrNoUserInContext.Error()})
		return
	}

	var req []requests.PoolEntryCreateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	entries, err := c.poolEntryService.CreateBatch(currentUser, req)
	if err != nil {
		log.Printf("LOG: (PoolEntryController: CreateBatch) - Service method error: %v\n", err)
		switch err {
		case types.ErrInvalidState:
			ctx.JSON(http.StatusForbidden, gin.H{"error": types.ErrInvalidState.Error()})
		case types.ErrLeagueNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrLeagueNotFound.Error()})
		case types.ErrPokemonSpeciesNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrPokemonSpeciesNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, entries)
}

// Update godoc
//
//	@Summary		Update a pool entry
//	@Description	Modify a pool entry's cost or availability
//	@Tags			Pool Entries
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string								true	"League ID"
//	@Param			request		body		requests.PoolEntryUpdateRequestDTO	true	"Updated entry"
//	@Success		200			{object}	models.PoolEntry
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		403			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/pool-entries [put]
func (c *poolEntryControllerImpl) Update(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrNoUserInContext.Error()})
		return
	}

	var req requests.PoolEntryUpdateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	entry, err := c.poolEntryService.Update(currentUser, &req)
	if err != nil {
		log.Printf("LOG: (PoolEntryController: Update) - Service method error: %v\n", err)
		switch err {
		case types.ErrPoolEntryNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrPoolEntryNotFound.Error()})
		case types.ErrLeagueNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrLeagueNotFound.Error()})
		case types.ErrInvalidState:
			ctx.JSON(http.StatusForbidden, gin.H{"error": types.ErrInvalidState.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, entry)
}
