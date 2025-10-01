package controllers

import (
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PokemonSpeciesController handles pokemon species related HTTP requests
// unprotected routes
type PokemonSpeciesController interface {
	// GET to get all pokemon
	GetAllPokemonSpecies(ctx *gin.Context)
}

type pokemonSpeciesControllerImpl struct {
	pokemonService services.PokemonSpeciesService
}

func NewPokemonSpeciesController(pokemonService services.PokemonSpeciesService) *pokemonSpeciesControllerImpl {
	return &pokemonSpeciesControllerImpl{
		pokemonService: pokemonService,
	}
}

// GET api/pokemon_species/
// get all pokemon species
func (c *pokemonSpeciesControllerImpl) GetAllPokemonSpecies(ctx *gin.Context) {
	pokemonDTOs, err := c.pokemonService.GetAllPokemonSpecies()
	if err != nil {
		log.Printf("(Error: PokemonSpeciesController.GetAllPokemonSpecies) - Service failed: %v\n", err)
		// The service method GetAllPokemonSpecies currently only returns common.ErrInternalService
		// if an error occurs. It does not return common.ErrPokemonSpeciesNotFound.
		// Therefore, we handle it as a generic internal server error.
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Pokemon species due to an internal error"})
		return
	}

	ctx.JSON(http.StatusOK, pokemonDTOs)
}
