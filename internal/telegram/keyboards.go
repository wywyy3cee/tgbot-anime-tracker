package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

// inline keyboards
func (b *Bot) createAnimeKeyboard(userID int64, animeID int, isFavorite bool, userRating *models.Rating) tgbotapi.InlineKeyboardMarkup {
	state := b.getState(userID)
	if state == nil {
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	var buttons [][]tgbotapi.InlineKeyboardButton

	if len(state.SearchResults) > 1 {
		navRow := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("⬅️", "prev"),
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%d/%d", state.CurrentIndex+1, len(state.SearchResults)),
				"position",
			),
			tgbotapi.NewInlineKeyboardButtonData("➡️", "next"),
		}
		buttons = append(buttons, navRow)
	}

	actionRow := []tgbotapi.InlineKeyboardButton{}

	if isFavorite {
		actionRow = append(actionRow, tgbotapi.NewInlineKeyboardButtonData("💔 Удалить", fmt.Sprintf("unfav:%d", animeID)))
	} else {
		actionRow = append(actionRow, tgbotapi.NewInlineKeyboardButtonData("❤️ Добавить", fmt.Sprintf("fav:%d", animeID)))
	}

	ratingText := "⭐ Оценить"
	if userRating != nil {
		ratingText = fmt.Sprintf("⭐ Оценка: %d", userRating.Score)
	}
	actionRow = append(actionRow, tgbotapi.NewInlineKeyboardButtonData(ratingText, fmt.Sprintf("rate:%d", animeID)))

	buttons = append(buttons, actionRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func (b *Bot) createRatingKeyboard(animeID int) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	row1 := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("1", fmt.Sprintf("rating:%d:1", animeID)),
		tgbotapi.NewInlineKeyboardButtonData("2", fmt.Sprintf("rating:%d:2", animeID)),
		tgbotapi.NewInlineKeyboardButtonData("3", fmt.Sprintf("rating:%d:3", animeID)),
		tgbotapi.NewInlineKeyboardButtonData("4", fmt.Sprintf("rating:%d:4", animeID)),
		tgbotapi.NewInlineKeyboardButtonData("5", fmt.Sprintf("rating:%d:5", animeID)),
	}

	row2 := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("6", fmt.Sprintf("rating:%d:6", animeID)),
		tgbotapi.NewInlineKeyboardButtonData("7", fmt.Sprintf("rating:%d:7", animeID)),
		tgbotapi.NewInlineKeyboardButtonData("8", fmt.Sprintf("rating:%d:8", animeID)),
		tgbotapi.NewInlineKeyboardButtonData("9", fmt.Sprintf("rating:%d:9", animeID)),
		tgbotapi.NewInlineKeyboardButtonData("10", fmt.Sprintf("rating:%d:10", animeID)),
	}

	cancelRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_rating"),
	}

	buttons = append(buttons, row1, row2, cancelRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func (b *Bot) createFavoritesKeyboard(favorites []models.Favorite, currentPage, totalPages int) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	start := currentPage * 10
	end := start + 10
	if end > len(favorites) {
		end = len(favorites)
	}

	for i := start; i < end; i++ {
		fav := favorites[i]
		title := fav.Title
		if len(title) > 60 {
			title = title[:57] + "..."
		}
		buttonData := fmt.Sprintf("show_fav:%d", fav.AnimeID)

		button := tgbotapi.NewInlineKeyboardButtonData(title, buttonData)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	if totalPages > 1 {
		navRow := []tgbotapi.InlineKeyboardButton{}

		if currentPage > 0 {
			navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("⬅️", "fav_prev"))
		}

		pageText := fmt.Sprintf("Стр. %d/%d", currentPage+1, totalPages)
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData(pageText, "fav_page"))

		if currentPage < totalPages-1 {
			navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("➡️", "fav_next"))
		}

		buttons = append(buttons, navRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func (b *Bot) createFavoriteAnimeKeyboard(animeID int, userRating *models.Rating) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	deleteRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить из избранного", fmt.Sprintf("del_fav:%d", animeID)),
	}
	buttons = append(buttons, deleteRow)

	ratingText := "⭐ Оценить"
	if userRating != nil {
		ratingText = fmt.Sprintf("⭐ Оценка: %d", userRating.Score)
	}
	ratingRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(ratingText, fmt.Sprintf("rate:%d", animeID)),
	}
	buttons = append(buttons, ratingRow)

	backRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад к списку", "back_to_favs"),
	}
	buttons = append(buttons, backRow)

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// reply keyboards
func (b *Bot) createMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Поиск"),
			tgbotapi.NewKeyboardButton("Избранное"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Помощь"),
		),
	)
}

func (b *Bot) createCancelKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Отмена"),
		),
	)
}
