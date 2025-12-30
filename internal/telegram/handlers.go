package telegram

import (
	"fmt"
	"math"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

const favoritesPerPage = 10

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

		b.handleSearch(userID, chatID, message.Text)
		return
	}

	if message.IsCommand() {
		switch message.Command() {
		case "start":
			b.handleStart(message)
		case "search":
			query := message.CommandArguments()
			if query != "" {
				b.handleSearch(userID, chatID, query)
			} else {
				b.handleSearchButton(userID, chatID)
			}
		case "favorites":
			b.handleFavorites(userID, chatID)
		}
		return
	}

	switch message.Text {
	case "Поиск":
		b.handleSearchButton(userID, chatID)
	case "Избранное":
		b.handleFavorites(userID, chatID)
	case "Помощь":
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
		"/favorites - избранное"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = b.createMainMenuKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := "Привет! Я бот для поиска аниме.\n\n" +
		"Команды:\n" +
		"/search <название> - поиск аниме\n" +
		"/favorites - твое избранное"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = b.createMainMenuKeyboard()
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
	state := b.getState(userID)
	if state != nil {
		state.WaitingForSearch = false
		b.saveState(userID, state)
	}

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

	state = &UserState{
		SearchResults: animes,
		CurrentIndex:  0,
	}
	b.saveState(userID, state)

	b.showCurrentAnime(chatID, userID)
}

func (b *Bot) handleFavorites(userID int64, chatID int64) {
	state := b.getState(userID)
	if state == nil {
		state = &UserState{FavoritesPage: 0}
		b.saveState(userID, state)
	}

	favorites, err := b.animeService.GetUserFavorites(userID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Ошибка получения избранного")
		msg.ReplyMarkup = b.createMainMenuKeyboard()
		b.api.Send(msg)
		return
	}

	if len(favorites) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Твое избранное пусто. Добавь аниме через поиск!")
		msg.ReplyMarkup = b.createMainMenuKeyboard()
		b.api.Send(msg)
		return
	}

	b.showFavoritesPage(chatID, userID, favorites)
}

func (b *Bot) showFavoritesPage(chatID int64, userID int64, favorites []models.Favorite) {
	state := b.getState(userID)
	if state == nil {
		state = &UserState{FavoritesPage: 0}
		b.saveState(userID, state)
	}

	totalPages := int(math.Ceil(float64(len(favorites)) / float64(favoritesPerPage)))
	currentPage := state.FavoritesPage

	if currentPage < 0 {
		currentPage = 0
	}
	if currentPage >= totalPages {
		currentPage = totalPages - 1
	}
	state.FavoritesPage = currentPage
	b.saveState(userID, state)

	text := fmt.Sprintf("❤️ Твое избранное (%d):\n\nВыбери аниме для просмотра:", len(favorites))

	keyboard := b.createFavoritesKeyboard(favorites, currentPage, totalPages)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) showCurrentAnime(chatID int64, userID int64) {
	anime := b.getCurrentAnime(userID)
	if anime == nil {
		msg := tgbotapi.NewMessage(chatID, "Аниме не найдено")
		msg.ReplyMarkup = b.createMainMenuKeyboard()
		b.api.Send(msg)
		return
	}

	isFav, _ := b.animeService.IsFavorite(userID, anime.ID)

	text := fmt.Sprintf(
		"🎬 *%s*\n%s\n\n"+
			"📺 Тип: %s\n"+
			"⭐ Оценка: %s\n"+
			"📊 Статус: %s\n"+
			"📺 Эпизодов: %d",
		anime.Name,
		anime.Russian,
		anime.Kind,
		anime.Score,
		anime.Status,
		anime.Episodes,
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

	b.logger.Debug("User %d clicked callback: %s", userID, data)

	switch data {
	case "next":
		state := b.getState(userID)
		if state != nil && len(state.SearchResults) > 0 {
			state.CurrentIndex++
			if state.CurrentIndex >= len(state.SearchResults) {
				state.CurrentIndex = 0
			}
			b.saveState(userID, state)
			b.editCurrentAnime(callback.Message.Chat.ID, callback.Message.MessageID, userID)
		}
		b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
		return

	case "prev":
		state := b.getState(userID)
		if state != nil && len(state.SearchResults) > 0 {
			state.CurrentIndex--
			if state.CurrentIndex < 0 {
				state.CurrentIndex = len(state.SearchResults) - 1
			}
			b.saveState(userID, state)
			b.editCurrentAnime(callback.Message.Chat.ID, callback.Message.MessageID, userID)
		}
		b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
		return

	case "position":
		b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
		return

	case "fav_next":
		state := b.getState(userID)
		if state != nil {
			state.FavoritesPage++
			b.saveState(userID, state)

			favorites, _ := b.animeService.GetUserFavorites(userID)
			b.editFavoritesPage(callback.Message.Chat.ID, callback.Message.MessageID, userID, favorites)
		}
		b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
		return

	case "fav_prev":
		state := b.getState(userID)
		if state != nil && state.FavoritesPage > 0 {
			state.FavoritesPage--
			b.saveState(userID, state)

			favorites, _ := b.animeService.GetUserFavorites(userID)
			b.editFavoritesPage(callback.Message.Chat.ID, callback.Message.MessageID, userID, favorites)
		}
		b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
		return

	case "fav_page":
		b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
		return

	case "back_to_favs":
		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		b.api.Send(deleteMsg)
		b.handleFavorites(userID, callback.Message.Chat.ID)
		b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	if len(data) > 9 && data[:9] == "show_fav:" {
		animeID := 0
		fmt.Sscanf(data, "show_fav:%d", &animeID)

		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		b.api.Send(deleteMsg)

		b.showFavoriteAnime(callback.Message.Chat.ID, userID, animeID)
		b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	if len(data) > 8 && data[:8] == "del_fav:" {
		animeID := 0
		fmt.Sscanf(data, "del_fav:%d", &animeID)

		err := b.animeService.RemoveFromFavorites(userID, animeID)
		if err != nil {
			b.logger.Error("Failed to delete from favorites: user %d, anime %d: %v", userID, animeID, err)
			b.api.Send(tgbotapi.NewCallback(callback.ID, "Ошибка удаления"))
			return
		}

		b.logger.Info("User %d deleted anime %d from favorites", userID, animeID)
		b.api.Send(tgbotapi.NewCallback(callback.ID, "💔 Удалено из избранного"))

		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		b.api.Send(deleteMsg)
		b.handleFavorites(userID, callback.Message.Chat.ID)
		return
	}

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

func (b *Bot) editFavoritesPage(chatID int64, messageID int, userID int64, favorites []models.Favorite) {
	state := b.getState(userID)
	if state == nil {
		return
	}

	totalPages := int(math.Ceil(float64(len(favorites)) / float64(favoritesPerPage)))
	currentPage := state.FavoritesPage

	if currentPage < 0 {
		currentPage = 0
	}
	if currentPage >= totalPages {
		currentPage = totalPages - 1
	}

	text := fmt.Sprintf("❤️ Твое избранное (%d):\n\nВыбери нужное аниме:", len(favorites))

	keyboard := b.createFavoritesKeyboard(favorites, currentPage, totalPages)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ReplyMarkup = &keyboard
	b.api.Send(edit)
}

func (b *Bot) showFavoriteAnime(chatID int64, userID int64, animeID int) {
	b.logger.Info("User %d viewing favorite anime ID: %d", userID, animeID)
	anime, err := b.animeService.GetAnimeByID(animeID)
	if err != nil {
		b.logger.Error("Failed to get anime details: %v", err)
		b.api.Send(tgbotapi.NewMessage(chatID, "Ошибка загрузки данных аниме"))
		return
	}

	text := fmt.Sprintf(
		"🎬 *%s*\n%s\n\n"+
			"📺 Тип: %s\n"+
			"⭐ Оценка: %s\n"+
			"📊 Статус: %s\n"+
			"📺 Эпизодов: %d\n\n"+
			"💚 В избранном",
		anime.Name,
		anime.Russian,
		anime.Kind,
		anime.Score,
		anime.Status,
		anime.Episodes,
	)

	if len(text) > 1024 {
		text = text[:1021] + "..."
	}

	keyboard := b.createFavoriteAnimeKeyboard(animeID)

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
