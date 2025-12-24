package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	username := message.From.UserName
	if username == "" {
		username = message.From.FirstName
	}
	b.animeService.EnsureUserExists(userID, username)

	state := b.getState(userID)
	if state != nil && state.WaitingForSearch {
		if message.Text == "Отмена" {
			state.WaitingForSearch = false
			b.saveState(userID, state)

			msg := tgbotapi.NewMessage(chatID, "Поиск отменен.")
			msg.ReplyMarkup = b.createMainMenuKeyboard()
			b.api.Send(msg)
			return
		}
	}

	if message.IsCommand() {
		switch message.Command() {
		case "start":
			b.handleStart(message)
		case "search":
			query := message.CommandArguments()
			b.handleSearch(userID, chatID, query)
		case "next":
			b.handleNext(userID, chatID)
		case "favorites":
			b.handleFavorites(userID, chatID)
		}
	}

	switch message.Text {
	case "🔍 Поиск":
		b.handleSearchButton(userID, chatID)
	case "❤️ Избранное":
		b.handleFavorites(userID, chatID)
	case "ℹ️ Помощь":
		b.handleHelp(message)
	default:
		msg := tgbotapi.NewMessage(chatID, "Используй кнопки меню или команды")
		msg.ReplyMarkup = b.createMainMenuKeyboard()
		b.api.Send(msg)
	}
}

func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := "ℹ️ Справка:\n\n" +
		"Поиск - найти аниме по названию\n" +
		"Избранное - список сохраненных аниме\n\n" +
		"Команды:\n" +
		"/search <название> - поиск\n" +
		"/next - следующее\n" +
		"/favorites - избранное"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = b.createMainMenuKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := "Привет! Я бот для поиска аниме.\n\n" +
		"Команды:\n" +
		"/search <название> - поиск аниме\n" +
		"/next - следующее аниме из результатов\n" +
		"/favorites - твое избранное"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)
}

func (b *Bot) handleSearchButton(userID int64, chatID int64) {
	state := b.getState(userID)
	if state == nil {
		state = &UserState{}
	}
	state.WaitingForSearch = true
	b.saveState(userID, state)

	msg := tgbotapi.NewMessage(chatID, "Напиши название аниме для поиска:")
	msg.ReplyMarkup = b.createCancelKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleSearch(userID int64, chatID int64, query string) {
	b.logger.Info("User %d searching for: %s", userID, query)

	if query == "" {
		msg := tgbotapi.NewMessage(chatID, "Укажи название. Например: /search bebop")
		msg.ReplyMarkup = b.createMainMenuKeyboard()
		b.api.Send(msg)
		return
	}

	animes, err := b.animeService.SearchAnime(query)
	if err != nil {
		b.logger.Error("Search failed for user %d, query '%s': %v", userID, query, err)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка: %v", err))
		msg.ReplyMarkup = b.createMainMenuKeyboard()
		b.api.Send(msg)
		return
	}

	b.logger.Info("Found %d animes for query '%s'", len(animes), query)

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
		msg := tgbotapi.NewMessage(chatID, "Это было последнее аниме. Начинаем сначала.")
		b.api.Send(msg)
		state.CurrentIndex = 0
	}

	b.saveState(userID, state)
	b.showCurrentAnime(chatID, userID)
}

// TODO:
// 1. сделать отдельной кнопкой переключение на следующую страницу избранного,
// 2. если делать полноценное отображение скорее всего придётся хранить данные в редисе???

func (b *Bot) handleFavorites(userID int64, chatID int64) {
	favorites, err := b.animeService.GetUserFavorites(userID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Ошибка получения избранного")
		b.api.Send(msg)
		return
	}

	if len(favorites) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Твое избранное пусто. Добавь аниме через поиск!")
		b.api.Send(msg)
		return
	}

	text := fmt.Sprintf("❤️ Твое избранное (%d):\n\n", len(favorites))
	for i, fav := range favorites {
		text += fmt.Sprintf("%d. %s\n", i+1, fav.Title)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	b.api.Send(msg)
}

func (b *Bot) showCurrentAnime(chatID int64, userID int64) {
	anime := b.getCurrentAnime(userID)
	if anime == nil {
		msg := tgbotapi.NewMessage(chatID, "Аниме не найдено")
		b.api.Send(msg)
		return
	}

	state := b.getState(userID)
	isFav, _ := b.animeService.IsFavorite(userID, anime.ID)

	text := fmt.Sprintf(
		"🎬 *%s*\n%s\n\n"+
			"📺 Тип: %s\n"+
			"⭐ Оценка: %s\n"+
			"📊 Статус: %s\n"+
			"📺 Эпизодов: %d\n\n"+
			"Показано %d из %d\n\n"+
			"Используй /next для следующего",
		anime.Name,
		anime.Russian,
		anime.Kind,
		anime.Score,
		anime.Status,
		anime.Episodes,
		state.CurrentIndex+1,
		len(state.SearchResults),
	)

	if isFav {
		text += "\n\n💚 В избранном"
	}

	if len(text) > 1024 {
		text = text[:1021] + "..."
	}

	keyboard := b.createAnimeKeyboard(userID, anime.ID, isFav)

	if anime.Image.Original != "" || anime.Image.Preview != "" {
		baseURL := "https://shikimori.one"
		imagePath := anime.Image.Original
		if imagePath == "" {
			imagePath = anime.Image.Preview
		}

		fullURL := baseURL + imagePath

		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(fullURL))
		photo.Caption = text
		photo.ParseMode = "Markdown"
		photo.ReplyMarkup = keyboard
		b.api.Send(photo)
	} else {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
	}
}

func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	data := callback.Data

	if len(data) > 4 && data[:4] == "fav:" {
		animeID := 0
		fmt.Sscanf(data, "fav:%d", &animeID)

		anime := b.getCurrentAnime(userID)
		if anime == nil {
			b.api.Send(tgbotapi.NewCallback(callback.ID, "Ошибка"))
			return
		}

		err := b.animeService.AddToFavorites(userID, *anime)
		if err != nil {
			b.logger.Error("Failed to add to favorites: user %d, anime %d: %v", userID, animeID, err)
			b.api.Send(tgbotapi.NewCallback(callback.ID, "Ошибка добавления"))
			return
		}

		b.logger.Info("User %d added anime %d to favorites", userID, animeID)
		b.api.Send(tgbotapi.NewCallback(callback.ID, "✅ Добавлено в избранное"))

		b.editCurrentAnime(callback.Message.Chat.ID, callback.Message.MessageID, userID)
		return
	}

	if len(data) > 6 && data[:6] == "unfav:" {
		animeID := 0
		fmt.Sscanf(data, "unfav:%d", &animeID)

		err := b.animeService.RemoveFromFavorites(userID, animeID)
		if err != nil {
			b.logger.Error("Failed to delete from favorites: user %d, anime %d: %v", userID, animeID, err)
			b.api.Send(tgbotapi.NewCallback(callback.ID, "Ошибка удаления"))
			return
		}

		b.logger.Info("User %d deleted anime %d from favorites", userID, animeID)
		b.api.Send(tgbotapi.NewCallback(callback.ID, "💔 Удалено из избранного"))

		b.editCurrentAnime(callback.Message.Chat.ID, callback.Message.MessageID, userID)
		return
	}
}

func (b *Bot) editCurrentAnime(chatID int64, messageID int, userID int64) {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	b.api.Send(deleteMsg)

	b.showCurrentAnime(chatID, userID)
}
