package controllers

import (
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
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
	logger           *slog.Logger
	poolEntryService services.PoolEntryService
}

func NewPoolEntryController(logger *slog.Logger, poolEntryService services.PoolEntryService) PoolEntryController {
	return &poolEntryControllerImpl{
		logger:           utils.LoggerWithService(logger, "PoolEntryController"),
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
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	entry, err := c.poolEntryService.GetByID(entryID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByID")
		switch err {
		case types.ErrPoolEntryNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrPoolEntryNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
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
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	entries, err := c.poolEntryService.GetByLeague(leagueID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByLeague")
		switch err {
		case types.ErrLeagueNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrLeagueNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
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
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	entries, err := c.poolEntryService.GetAvailableByLeague(leagueID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetAvailableByLeague")
		switch err {
		case types.ErrLeagueNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrLeagueNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
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
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	var req requests.PoolEntryCreateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "bad request")
		return
	}

	entry, err := c.poolEntryService.Create(currentUser, &req)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "Create")
		switch err {
		case types.ErrInvalidState:
			sendError(ctx, http.StatusForbidden, types.ErrInvalidState.Error())
		case types.ErrLeagueNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrLeagueNotFound.Error())
		case types.ErrPokemonSpeciesNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrPokemonSpeciesNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
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
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	var req []requests.PoolEntryCreateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "bad request")
		return
	}

	entries, err := c.poolEntryService.CreateBatch(currentUser, req)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "CreateBatch")
		switch err {
		case types.ErrInvalidState:
			sendError(ctx, http.StatusForbidden, types.ErrInvalidState.Error())
		case types.ErrLeagueNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrLeagueNotFound.Error())
		case types.ErrPokemonSpeciesNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrPokemonSpeciesNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
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
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	var req requests.PoolEntryUpdateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "bad request")
		return
	}

	entry, err := c.poolEntryService.Update(currentUser, &req)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "Update")
		switch err {
		case types.ErrPoolEntryNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrPoolEntryNotFound.Error())
		case types.ErrLeagueNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrLeagueNotFound.Error())
		case types.ErrInvalidState:
			sendError(ctx, http.StatusForbidden, types.ErrInvalidState.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, entry)
}
