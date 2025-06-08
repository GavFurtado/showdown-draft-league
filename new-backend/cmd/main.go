package main

import (
	"os"

	routes "github.com/GavFurtado/showdown-draft-league/new-backend/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := gin.Default()

	// middleware goes here

	routes.RegisterRoutes(server)

	server.Run(":" + port)
}
