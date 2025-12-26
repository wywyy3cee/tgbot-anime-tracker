package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// inline button's search
func (b *Bot) createAnimeKeyboard(userID int64, animeID int, isFavorite bool) tgbotapi.InlineKeyboardMarkup {
	state := b.getState(userID)
	if state == nil {
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	var buttons [][]tgbotapi.InlineKeyboardButton

	if len(state.SearchResults) > 1 {
		navRow := []tgbotapi.InlineKeyboardButton{}
		if state.CurrentIndex > 0 {
			navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("⬅️", "prev"))
		}
		positionText := fmt.Sprintf(" %d/%d ", state.CurrentIndex+1, len(state.SearchResults))
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData(positionText, "position"))
		if state.CurrentIndex < len(state.SearchResults)-1 {
			navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("➡️", "next"))
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
func (b *Bot) createFavoritesKeyboard(currentPage, totalPages int) tgbotapi.InlineKeyboardMarkup {
	if totalPages <= 1 {
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	navRow := []tgbotapi.InlineKeyboardButton{}

	if currentPage > 0 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("⬅️", "fav_prev"))
	}

	pageText := fmt.Sprintf("Стр. %d/%d", currentPage+1, totalPages)
	navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData(pageText, "fav_page"))

	if currentPage < totalPages-1 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("➡️", "fav_next"))
	}
	return tgbotapi.NewInlineKeyboardMarkup(navRow)
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
