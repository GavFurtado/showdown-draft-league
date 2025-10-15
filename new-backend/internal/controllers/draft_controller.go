package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DraftController interface {
	GetDraftByID(ctx *gin.Context)
	GetDraftByLeagueID(ctx *gin.Context)
	StartDraft(ctx *gin.Context)
	StartTransferPeriod(ctx *gin.Context)
	EndTransferPeriod(ctx *gin.Context)
	MakePick(ctx *gin.Context)
	SkipPick(ctx *gin.Context)
}

type draftControllerImpl struct {
	draftService services.DraftService
}

func NewDraftController(draftService services.DraftService) *draftControllerImpl {
	return &draftControllerImpl{
		draftService: draftService,
	}
}

func (dc *draftControllerImpl) GetDraftByID(ctx *gin.Context) {
	draftIDStr := ctx.Param("draftId")
	draftID, err := uuid.Parse(draftIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	draft, err := dc.draftService.GetDraftByID(draftID)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrDraftNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrDraftNotFound.Error()})
		case errors.Is(err, common.ErrInternalService):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error(), "details": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, draft)

}

func (dc *draftControllerImpl) GetDraftByLeagueID(ctx *gin.Context) {
	leagueIDStr := ctx.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	draft, err := dc.draftService.GetDraftByLeagueID(leagueID)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrDraftNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrDraftNotFound.Error()})
		case errors.Is(err, common.ErrInternalService):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

func (dc *draftControllerImpl) StartDraft(ctx *gin.Context) {
	leagueIDStr := ctx.Param("leagueId")

	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	turnTimeLimitStr := ctx.DefaultQuery("turnTimeLimit", "120") // Default to 120 minutes
	turnTimeLimit, err := strconv.Atoi(turnTimeLimitStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid turnTimeLimit value"})
		return
	}

	draft, err := dc.draftService.StartDraft(leagueID, turnTimeLimit)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrLeagueNotFound):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrLeagueNotFound.Error()})
		case errors.Is(err, common.ErrNoPlayerForDraft):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrNoPlayerForDraft.Error()})
		case errors.Is(err, common.ErrInvalidDraftPosition):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInvalidDraftPosition.Error()})
		case errors.Is(err, common.ErrDuplicateDraftPosition):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrDuplicateDraftPosition.Error()})
		case errors.Is(err, common.ErrIncompleteDraftOrder):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrIncompleteDraftOrder.Error()})
		case errors.Is(err, common.ErrInternalService):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrIncompleteDraftOrder.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start draft", "details": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

// StartTransferPeriod handles the POST /api/leagues/:leagueId/transfers/start endpoint.
// It initiates the transfer window for a specific league.
func (dc *draftControllerImpl) StartTransferPeriod(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	if err := dc.draftService.StartTransferPeriod(leagueID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transfer period", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer period started successfully"})
}

func (dc *draftControllerImpl) EndTransferPeriod(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	if err := dc.draftService.EndTransferPeriod(leagueID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end transfer period", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer period ended successfully"})
}

func (dc *draftControllerImpl) MakePick(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	var input common.DraftMakePickDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ErrInvalidInput.Error()})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrNoUserInContext.Error()})
		return
	}
	user, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user information"})
		return
	}

	if err := dc.draftService.MakePick(user, leagueID, &input); err != nil {
		// Handle specific errors from the service layer
		switch {
		case errors.Is(err, common.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not your turn to pick"})
		case errors.Is(err, common.ErrInvalidState):
			c.JSON(http.StatusConflict, gin.H{"error": "Draft is not in a valid state for picking"})
		case errors.Is(err, common.ErrTooManyRequestedPicks):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Requested too many picks"})
		case errors.Is(err, common.ErrInsufficientDraftPoints):
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient draft points"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make pick", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pick successful"})
}

func (dc *draftControllerImpl) SkipPick(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user information"})
		return
	}

	if err := dc.draftService.SkipTurn(user, leagueID); err != nil {
		switch {
		case errors.Is(err, common.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not your turn to skip"})
		case errors.Is(err, common.ErrInvalidState):
			c.JSON(http.StatusConflict, gin.H{"error": "Draft is not in a valid state for skipping"})
		case errors.Is(err, common.ErrCannotSkipBelowMinimumRoster):
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot skip, minimum roster requirement not met"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to skip turn", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Turn skipped successfully"})
}
