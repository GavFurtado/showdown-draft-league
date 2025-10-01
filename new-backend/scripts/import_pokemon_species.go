package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os" // Only os is needed now
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	newLogger := logger.New(
		log.New(os.Stdout, "\\r\\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Ensure the PokemonSpecies table exists and is migrated
	if !db.Migrator().HasTable(&models.PokemonSpecies{}) {
		log.Println("Creating PokemonSpecies table...")
		err = db.Migrator().CreateTable(&models.PokemonSpecies{})
		if err != nil {
			log.Fatalf("Failed to create PokemonSpecies table: %v", err)
		}
	} else {
		log.Println("Migrating PokemonSpecies table...")
		err = db.AutoMigrate(&models.PokemonSpecies{})
		if err != nil {
			log.Fatalf("Failed to auto-migrate PokemonSpecies table: %v", err)
		}
	}

	// CLI menu option to delete existing records
	fmt.Print("Delete all existing PokemonSpecies records before import? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" {
		log.Println("Deleting all existing PokemonSpecies records...")
		deleteResult := db.Exec("DELETE FROM pokemon_species")
		if deleteResult.Error != nil {
			log.Fatalf("Failed to delete existing PokemonSpecies records: %v", deleteResult.Error)
		}
		log.Printf("Deleted %d existing PokemonSpecies records.", deleteResult.RowsAffected)
	}

	pokemonRepo := repositories.NewPokemonSpeciesRepository(db)

	// Read the JSON file
	jsonFilePath := "data/all_pokemon_data.json"
	byteValue, err := os.ReadFile(jsonFilePath) // Use os.ReadFile
	if err != nil {
		log.Fatalf("Failed to read JSON file %s: %v", jsonFilePath, err)
	}

	var pokemonSpeciesList []models.PokemonSpecies
	err = json.Unmarshal(byteValue, &pokemonSpeciesList)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	log.Printf("Found %d Pokemon species in JSON file. Starting import...", len(pokemonSpeciesList))

	// Import data
	for _, pokemon := range pokemonSpeciesList {
		// Check if Pokemon already exists by ID
		existingPokemon, err := pokemonRepo.GetPokemonSpeciesByID(pokemon.ID)
		if err != nil && err.Error() != fmt.Sprintf("pokemon species with ID %d not found", pokemon.ID) {
			log.Printf("Error checking existing Pokemon with ID %d (DexID: %d): %v", pokemon.ID, pokemon.DexID, err)
			continue
		}

		if existingPokemon != nil {
			log.Printf("Pokemon species with ID %d (DexID: %d, %s) already exists, skipping.", pokemon.ID, pokemon.DexID, pokemon.Name)
			continue
		}

		// Create new Pokemon species
		err = pokemonRepo.CreatePokemonSpecies(&pokemon)
		if err != nil {
			log.Printf("Error creating Pokemon species ID %d (DexID: %d, %s): %v", pokemon.ID, pokemon.DexID, pokemon.Name, err)
			continue
		}
		log.Printf("Successfully imported Pokemon species ID %d (DexID: %d): %s", pokemon.ID, pokemon.DexID, pokemon.Name)
	}

	log.Println("Pokemon species import complete.")
}
