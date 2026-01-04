package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/cache"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/config"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/database"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/service"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/shikimori"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/telegram"
	"github.com/wywyy3cee/tgbot-anime-tracker/pkg/logger"
)

func main() {
	appLogger := logger.New()

	if os.Getenv("RAILWAY_ENVIRONMENT_NAME") == "" {
		err := godotenv.Load()
		if err != nil {
			appLogger.Info(".env file not found, using environment variables")
		}
	}

	cfg, err := config.Load()
	if err != nil {
		appLogger.Error("Failed to load config: %v", err)
		log.Fatal(err)
	}

	appLogger.Info("Starting application...")
	db, err := database.Connect(cfg.DatabaseURL, appLogger)
	if err != nil {
		appLogger.Error("Failed to connect to database: %v", err)
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.RunMigrations("./migrations"); err != nil {
		appLogger.Error("Failed to run migrations: %v", err)
		log.Fatal(err)
	}

	appLogger.Info("Database connected and migrations applied")

	log.Println("Database connected and migrations applied")

	redisCache, err := cache.New(cfg.RedisURL, appLogger)
	if err != nil {
		appLogger.Error("Failed to connect to redis: %v", err)
		log.Fatal(err)
	}
	defer redisCache.Close()

	repo := database.NewRepository(db)
	shikiClient := shikimori.NewClient(cfg.ShikimoriURL)
	animeService := service.NewAnimeService(shikiClient, repo, redisCache)

	bot, err := telegram.NewBot(cfg.BotToken, animeService, appLogger)
	if err != nil {
		appLogger.Error("Failed to create bot: %v", err)
		log.Fatal("Failed to create bot:", err)
	}

	log.Printf("Bot started successfully")

	if err := bot.Start(); err != nil {
		appLogger.Error("Bot stopped with error: %v", err)
		log.Fatal("Bot stopped:", err)
	}
}
