package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

// inline button's search
func (b *Bot) createAnimeKeyboard(userID int64, animeID int, isFavorite bool) tgbotapi.InlineKeyboardMarkup {
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

	if isFavorite {
		unfavBtn := tgbotapi.NewInlineKeyboardButtonData("💔 Удалить из избранного", fmt.Sprintf("unfav:%d", animeID))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{unfavBtn})
	} else {
		favBtn := tgbotapi.NewInlineKeyboardButtonData("❤️ Добавить в избранное", fmt.Sprintf("fav:%d", animeID))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{favBtn})
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// inline button's favourites
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

func (b *Bot) createFavoriteAnimeKeyboard(animeID int) tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить из избранного", fmt.Sprintf("del_fav:%d", animeID)),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад к списку", "back_to_favs"),
		},
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// reply button's main menu
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
