package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DraftPickController interface {
	GetByDraft(ctx *gin.Context)
	GetByPlayer(ctx *gin.Context)
	GetHistory(ctx *gin.Context)
	GetNextPickNumber(ctx *gin.Context)
}

type draftPickControllerImpl struct {
	draftPickService services.DraftPickService
	draftService     services.DraftService
}

func NewDraftPickController(
	draftPickService services.DraftPickService,
	draftService services.DraftService,
) DraftPickController {
	return &draftPickControllerImpl{
		draftPickService: draftPickService,
		draftService:     draftService,
	}
}

func (c *draftPickControllerImpl) GetByDraft(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	draft, err := c.draftService.GetDraftByLeagueID(leagueID)
	if err != nil {
		log.Printf("LOG: (DraftPickController: GetByDraft) - failed to get draft: %v\n", err)
		switch {
		case errors.Is(err, types.ErrDraftNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrDraftNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	picks, err := c.draftPickService.GetByDraft(draft.ID)
	if err != nil {
		log.Printf("LOG: (DraftPickController: GetByDraft) - Service method error: %v\n", err)
		switch {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, picks)
}

func (c *draftPickControllerImpl) GetByPlayer(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	picks, err := c.draftPickService.GetByPlayer(playerID)
	if err != nil {
		log.Printf("LOG: (DraftPickController: GetByPlayer) - Service method error: %v\n", err)
		switch {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, picks)
}

func (c *draftPickControllerImpl) GetHistory(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	history, err := c.draftPickService.GetHistory(leagueID)
	if err != nil {
		log.Printf("LOG: (DraftPickController: GetHistory) - Service method error: %v\n", err)
		switch {
		case errors.Is(err, types.ErrDraftNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrDraftNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, history)
}

func (c *draftPickControllerImpl) GetNextPickNumber(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	draft, err := c.draftService.GetDraftByLeagueID(leagueID)
	if err != nil {
		log.Printf("LOG: (DraftPickController: GetNextPickNumber) - failed to get draft: %v\n", err)
		switch {
		case errors.Is(err, types.ErrDraftNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrDraftNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	nextPickNumber, err := c.draftPickService.GetNextPickNumber(draft.ID)
	if err != nil {
		log.Printf("LOG: (DraftPickController: GetNextPickNumber) - Service method error: %v\n", err)
		switch {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"next_pick_number": nextPickNumber})
}
