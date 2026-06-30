package config

import (
	. "github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	ENVIRONMENT           string
	DISCORD_CLIENT_ID     string
	DISCORD_CLIENT_SECRET string
	DISCORD_REDIRECT_URI  string
	APP_BASE_URL          string
	BACKEND_BASE_URL      string
	DATABASE_URL          string
	JWT_SECRET            string
	PORT                  string
}

func LoadConfig() *Config {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	cfg := &Config{
		ENVIRONMENT:           getEnv("ENV"),
		DISCORD_CLIENT_ID:     getEnv("DISCORD_CLIENT_ID"),
		DISCORD_CLIENT_SECRET: getEnv("DISCORD_CLIENT_SECRET"),
		DISCORD_REDIRECT_URI:  getEnv("DISCORD_REDIRECT_URI"),
		APP_BASE_URL:          getEnv("APP_BASE_URL"),
		BACKEND_BASE_URL:      getEnv("BACKEND_BASE_URL"),
		DATABASE_URL:          getEnv("DATABASE_URL"),
		JWT_SECRET:            getEnv("JWT_SECRET"),
		PORT:                  getEnv("PORT"),
	}

	Assert(!(cfg.DISCORD_CLIENT_ID == "" || cfg.DISCORD_CLIENT_SECRET == "" || cfg.DISCORD_REDIRECT_URI == "" ||
		cfg.APP_BASE_URL == "" || cfg.DATABASE_URL == "" || cfg.JWT_SECRET == ""),
		"Missing required environment variables for Discord OAuth or database connection.")

	return cfg
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}
