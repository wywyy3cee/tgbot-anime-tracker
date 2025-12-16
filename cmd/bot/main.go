package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/database"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/shikimori"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/telegram"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := database.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatal(err)
	}

	log.Println("Database connected and migrations applied")

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN not found in environment variables")
	}

	shikiClient := shikimori.NewClient("https://shikimori.one/api")

	bot, err := telegram.NewBot(botToken, shikiClient)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	log.Printf("Bot started successfully")

	if err := bot.Start(); err != nil {
		log.Fatal("Bot stopped:", err)
	}
}
