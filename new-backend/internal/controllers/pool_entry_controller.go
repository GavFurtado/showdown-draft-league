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
