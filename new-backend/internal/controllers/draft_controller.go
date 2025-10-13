package controllers

import (
	"net/http"
	"strconv"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DraftController struct {
	draftService services.DraftService
	rbacService  services.RBACService
}

func NewDraftController(draftService services.DraftService, rbacService services.RBACService) *DraftController {
	return &DraftController{
		draftService: draftService,
		rbacService:  rbacService,
	}
}

func (dc *DraftController) StartDraft(c *gin.Context) {
	leagueIDStr := c.Param("league_id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	turnTimeLimitStr := c.DefaultQuery("turnTimeLimit", "120") // Default to 120 minutes
	turnTimeLimit, err := strconv.Atoi(turnTimeLimitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid turnTimeLimit value"})
		return
	}

	draft, err := dc.draftService.StartDraft(leagueID, turnTimeLimit)
	if err != nil {
		// More specific error handling can be added here based on the error type
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start draft", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, draft)
}

func (dc *DraftController) StartTransferPeriod(c *gin.Context) {
	leagueIDStr := c.Param("league_id")
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

func (dc *DraftController) EndTransferPeriod(c *gin.Context) {
	leagueIDStr := c.Param("league_id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	if err := dc.draftService.EndTransferPeriod(leagueID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end transfer period", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer period ended successfully"})
}
