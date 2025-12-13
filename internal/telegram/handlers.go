package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	if message.IsCommand() {
		switch message.Command() {
		case "start":
			b.handleStart(message)
		case "search":
			query := message.CommandArguments()
			b.handleSearch(userID, message.Chat.ID, query)
		case "next":
			b.handleNext(userID, chatID)
		}
	}
}

func (b *Bot) handleStart(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Хай, используй /search <название> для поиска тайтлов")
	b.api.Send(msg)
}

func (b *Bot) handleSearch(userID int64, chatID int64, query string) {
	if query == "" {
		msg := tgbotapi.NewMessage(chatID, "Укажи название аниме. Например /search lain")
		b.api.Send(msg)
		return
	}

	animes, err := b.shikimoriClient.SearchAnime(query, 10)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("error of searching: %v", err))
		b.api.Send(msg)
		return
	}

	state := &UserState{
		SearchResults: animes,
		CurrentIndex:  0,
	}
	b.saveState(userID, state)

	b.showCurrentAnime(chatID, userID)
}

func (b *Bot) handleNext(userID int64, chatID int64) {
	state := b.getState(userID)
	if state == nil {
		msg := tgbotapi.NewMessage(chatID, "Сначала сделай поиск: /search <название>")
		b.api.Send(msg)
		return
	}

	state.CurrentIndex++

	if state.CurrentIndex >= len(state.SearchResults) {
		msg := tgbotapi.NewMessage(chatID, "Начинаем сначала.")
		b.api.Send(msg)
		state.CurrentIndex = 0
	}

	b.saveState(userID, state)
	b.showCurrentAnime(chatID, userID)

}

func (b *Bot) showCurrentAnime(chatID int64, userID int64) {
	anime := b.getCurrentAnime(userID)
	if anime == nil {
		msg := tgbotapi.NewMessage(chatID, "Аниме не найдено!")
		b.api.Send(msg)
		return
	}

	state := b.getState(userID)

	text := fmt.Sprintf(
		"🎬 *%s*\n%s\n\n"+
			"Тип: %s\n"+
			"Оценка: %s\n"+
			"Статус: %s\n"+
			"Эпизодов: %d\n\n"+
			"Показано %d из %d",
		anime.Name,
		anime.Russian,
		anime.Kind,
		anime.Score,
		anime.Status,
		anime.Episodes,
		state.CurrentIndex+1,
		len(state.SearchResults),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}
