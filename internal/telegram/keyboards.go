package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) createAnimeKeyboard(userID int64, animeID int, isFavorite bool) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	if isFavorite {
		unfavBtn := tgbotapi.NewInlineKeyboardButtonData("💔 Удалить из избранного", fmt.Sprintf("unfav:%d", animeID))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{unfavBtn})
	} else {
		favBtn := tgbotapi.NewInlineKeyboardButtonData("❤️ Добавить в избранное", fmt.Sprintf("fav:%d", animeID))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{favBtn})
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// TODO: реплай кнопки
