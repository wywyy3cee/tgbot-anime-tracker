package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/database"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/service"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/shikimori"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/telegram"
	"github.com/wywyy3cee/tgbot-anime-tracker/pkg/logger"
)

func main() {
	appLogger := logger.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	appLogger.Info("Starting application...")
	db, err := database.Connect(os.Getenv("DATABASE_URL"), appLogger)
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

	repo := database.NewRepository(db)
	shikiClient := shikimori.NewClient("https://shikimori.one/api")
	animeService := service.NewAnimeService(shikiClient, repo)

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		appLogger.Error("BOT_TOKEN not found")
		log.Fatal("BOT_TOKEN not found in environment variables")
	}

	bot, err := telegram.NewBot(botToken, animeService, appLogger)
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
