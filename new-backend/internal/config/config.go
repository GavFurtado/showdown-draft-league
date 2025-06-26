package config

import (
	. "github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"log"
	"os"
)

type Config struct {
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURI  string
	AppBaseURL          string
	DatabaseURL         string
	JWTSecret           string
}

func LoadConfig() *Config {

	cfg := &Config{
		DiscordClientID:     getEnv("DISCORD_CLIENT_ID"),
		DiscordClientSecret: getEnv("DISCORD_CLIENT_SECRET"),
		DiscordRedirectURI:  getEnv("DISCORD_REDIRECT_URI"),
		AppBaseURL:          getEnv("APP_BASE_URL"),
		DatabaseURL:         getEnv("DATABASE_URL"),
		JWTSecret:           getEnv("JWT_SECRET"),
	}

	Assert((cfg.DiscordClientID == "" || cfg.DiscordClientSecret == "" || cfg.DiscordRedirectURI == "" ||
		cfg.AppBaseURL == "" || cfg.DatabaseURL == "" || cfg.JWTSecret == ""),
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
