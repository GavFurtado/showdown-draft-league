package controllers

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/responses"
	"github.com/gin-gonic/gin"
)

func sendError(c *gin.Context, status int, message string) {
	c.JSON(status, responses.NewErrorResponse(status, message, c.Request.URL.Path))
}
