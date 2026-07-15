package controllers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
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
	logger         *slog.Logger
	pokemonService services.PokemonSpeciesService
}

func NewPokemonSpeciesController(logger *slog.Logger, pokemonService services.PokemonSpeciesService) PokemonSpeciesController {
	return &pokemonSpeciesControllerImpl{
		logger:         utils.LoggerWithService(logger, "PokemonSpeciesController"),
		pokemonService: pokemonService,
	}
}

// GetAllPokemonSpecies godoc
//
//	@Summary		Get all Pokemon
//	@Description	Full list of available Pokemon species
//	@Tags			Pokemon
//	@Produce		json
//	@Success		200	{array}		responses.PokemonSpeciesListResponseDTO
//	@Failure		500	{object}	responses.ErrorResponse
//	@Security		none
//	@Router			/api/pokemon_species [get]
func (c *pokemonSpeciesControllerImpl) GetAllPokemonSpecies(ctx *gin.Context) {
	pokemonDTOs, err := c.pokemonService.GetAllPokemonSpecies()
	if err != nil {
		c.logger.Error("Service failed", "error", err, "method", "GetAllPokemonSpecies")
		sendError(ctx, http.StatusInternalServerError, "Failed to retrieve Pokemon species due to an internal error")
		return
	}

	ctx.JSON(http.StatusOK, pokemonDTOs)
}

// GetPokemonSpeciesByID godoc
//
//	@Summary		Get a Pokemon by ID
//	@Description	Pokemon species by internal ID
//	@Tags			Pokemon
//	@Produce		json
//	@Param			id	path		int	true	"Pokemon ID"
//	@Success		200	{object}	models.PokemonSpecies
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Security		none
//	@Router			/api/pokemon_species/{id} [get]
func (c *pokemonSpeciesControllerImpl) GetPokemonSpeciesByID(ctx *gin.Context) {
	pokemonIDstr := ctx.Param("id")
	pokemonID, err := strconv.ParseInt(pokemonIDstr, 10, 64)
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	pokemon, err := c.pokemonService.GetPokemonSpeciesByID(pokemonID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetPokemonSpeciesByID")
		switch err {
		case types.ErrPokemonSpeciesNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrPokemonSpeciesNotFound.Error())
			return
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
			return
		}
	}

	ctx.JSON(http.StatusOK, pokemon)
}

// GetPokemonSpeciesByName godoc
//
//	@Summary		Get a Pokemon by name
//	@Description	Pokemon species by name
//	@Tags			Pokemon
//	@Produce		json
//	@Param			name	path		string	true	"Pokemon name"
//	@Success		200		{object}	models.PokemonSpecies
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Security		none
//	@Router			/api/pokemon_species/name/{name} [get]
func (c *pokemonSpeciesControllerImpl) GetPokemonSpeciesByName(ctx *gin.Context) {
	pokemonName := ctx.Param("name")
	if pokemonName == "" {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	pokemon, err := c.pokemonService.GetPokemonSpeciesByName(pokemonName)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetPokemonSpeciesByName")
		switch err {
		case types.ErrPokemonSpeciesNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrPokemonSpeciesNotFound.Error())
			return
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
			return
		}
	}
	ctx.JSON(http.StatusOK, pokemon)
}
