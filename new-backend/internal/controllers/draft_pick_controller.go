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

// GetByDraft godoc
//
//	@Summary		Get draft picks
//	@Description	All picks made in a draft
//	@Tags			Draft Picks
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.DraftPick
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/draft-picks [get]
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

// GetByPlayer godoc
//
//	@Summary		Get picks by player
//	@Description	Draft picks for a specific player
//	@Tags			Draft Picks
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			playerId	path		string	true	"Player ID"
//	@Success		200			{array}		models.DraftPick
//	@Failure		400			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/draft-picks/player/{playerId} [get]
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

// GetHistory godoc
//
//	@Summary		Get draft pick history
//	@Description	Full history of all draft picks
//	@Tags			Draft Picks
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.DraftPick
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/draft-picks/history [get]
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

// GetNextPickNumber godoc
//
//	@Summary		Get next pick number
//	@Description	The next pick number in the draft
//	@Tags			Draft Picks
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/draft-picks/next-pick-number [get]
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
