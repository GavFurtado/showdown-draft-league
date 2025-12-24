package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PokemonSpeciesController handles pokemon species related HTTP requests
// unprotected routes
type PokemonSpeciesController interface {
	// GET to get all pokemon
	GetAllPokemonSpecies(ctx *gin.Context)
	// GET a pokemon species by it's ID
	GetPokemonSpeciesByID(ctx *gin.Context)
	// GET a pokemon species by it's name
	GetPokemonSpeciesByName(ctx *gin.Context)

	// admin only routes; not implemented
	// TODO: implement the admin only routes after admin only middleware checking is done
	//
	// // POST to create a new PokemonSpecies
	// CreatePokemonSpecies(ctx *gin.Context)
	// // PUT to update an existing PokemonSpecies
	// UpdatePokemonSpecies(ctx *gin.Context)
	// // DELETE an existing PokemonSpecies
	// DeletePokemonSpecies(ctx *gin.Context)
}

type pokemonSpeciesControllerImpl struct {
	pokemonService services.PokemonSpeciesService
}

func NewPokemonSpeciesController(pokemonService services.PokemonSpeciesService) PokemonSpeciesController {
	return &pokemonSpeciesControllerImpl{
		pokemonService: pokemonService,
	}
}

// GET api/pokemon_species/
// GetAllPokemonSpecies get all pokemon species
func (c *pokemonSpeciesControllerImpl) GetAllPokemonSpecies(ctx *gin.Context) {
	pokemonDTOs, err := c.pokemonService.GetAllPokemonSpecies()
	if err != nil {
		log.Printf("LOG: (Error: PokemonSpeciesController.GetAllPokemonSpecies) - Service failed: %v\n", err)
		// The service method GetAllPokemonSpecies currently only returns common.ErrInternalService
		// if an error occurs. It does not return common.ErrPokemonSpeciesNotFound.
		// Therefore, we handle it as a generic internal server error.
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Pokemon species due to an internal error"})
		return
	}

	ctx.JSON(http.StatusOK, pokemonDTOs)
}

// GET api/pokemon_species/:id
// GetPokemonSpeciesByID returns a PokemonSpecies by it's internal ID
// NOTE: this currently
func (c *pokemonSpeciesControllerImpl) GetPokemonSpeciesByID(ctx *gin.Context) {
	pokemonIDstr := ctx.Param("id")
	pokemonID, err := strconv.ParseInt(pokemonIDstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	pokemon, err := c.pokemonService.GetPokemonSpeciesByID(pokemonID)
	if err != nil {
		log.Printf("LOG: (PokemonSpeciesController: GetPokemonSpeciesByID): Service method Error: ")
		switch err {
		case common.ErrPokemonSpeciesNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrPokemonSpeciesNotFound.Error()})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
			return
		}
	}

	// Success
	ctx.JSON(http.StatusOK, pokemon)
}

// GET api/pokemon_species/name/:name
// GetPokemonSpeciesByID returns a PokemonSpecies by it's name
func (c *pokemonSpeciesControllerImpl) GetPokemonSpeciesByName(ctx *gin.Context) {
	pokemonName := ctx.Param("name")
	if pokemonName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	pokemon, err := c.pokemonService.GetPokemonSpeciesByName(pokemonName)
	if err != nil {
		log.Printf("LOG: (PokemonSpeciesController: GetPokemonSpeciesByName): Service method Error: ")
		switch err {
		case common.ErrPokemonSpeciesNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrPokemonSpeciesNotFound.Error()})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
			return
		}
	}
	// Success
	ctx.JSON(http.StatusOK, pokemon)
}
