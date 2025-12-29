package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL  string
	BotToken     string
	RedisURL     string
	ShikimoriURL string
}

func Load() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	shikimoriURL := os.Getenv("SHIKIMORI_URL")
	if shikimoriURL == "" {
		return nil, fmt.Errorf("SHIKIMORI_URL is required")
	}

	return &Config{
		DatabaseURL:  databaseURL,
		BotToken:     botToken,
		RedisURL:     redisURL,
		ShikimoriURL: shikimoriURL,
	}, nil
}
