package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

func TruncateTextWithEllipsis(text string, maxLength int) string {
	if len(text) > maxLength {
		return text[:maxLength-3] + "..."
	}
	return text
}

func SanitizeUTF8(text string) string {
	if utf8.ValidString(text) {
		return text
	}
	validRunes := []rune{}
	for _, r := range text {
		if r == utf8.RuneError {
			continue
		}
		validRunes = append(validRunes, r)
	}
	return string(validRunes)
}

func EscapeMarkdown(text string) string {
	result := text

	result = strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"~", "\\~",
		"`", "\\`",
	).Replace(result)

	result = removeJapaneseCharacters(result)

	result = removeBracketContent(result)

	return strings.TrimSpace(result)
}

func removeJapaneseCharacters(text string) string {
	validRunes := []rune{}
	for _, r := range text {
		if !isJapanese(r) {
			validRunes = append(validRunes, r)
		}
	}
	return string(validRunes)
}

func isJapanese(r rune) bool {
	return (r >= 0x3040 && r <= 0x309F) ||
		(r >= 0x30A0 && r <= 0x30FF) ||
		(r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF)
}

func removeBracketContent(text string) string {
	result := ""
	depth := 0
	for _, r := range text {
		if r == '[' {
			depth++
		} else if r == ']' {
			depth--
		} else if depth == 0 {
			result += string(r)
		}
	}
	return result
}

func FormatAnimeMessage(anime *models.Anime, isFav bool) string {
	description := TruncateTextWithEllipsis(anime.Description, 1024)
	description = SanitizeUTF8(description)
	description = EscapeMarkdown(description)
	descText := description
	if len(descText) == 0 {
		descText = "нет"
	}

	name := EscapeMarkdown(anime.Name)
	russian := EscapeMarkdown(anime.Russian)
	kind := EscapeMarkdown(anime.Kind)
	score := EscapeMarkdown(anime.Score)
	status := EscapeMarkdown(anime.Status)

	text := fmt.Sprintf(
		"🎬 %s\n%s\n\n"+
			"📺 Тип: %s\n"+
			"⭐ Оценка: %s\n"+
			"📊 Статус: %s\n"+
			"📺 Эпизодов: %d\n\n"+
			"Описание: %s",
		name,
		russian,
		kind,
		score,
		status,
		anime.Episodes,
		descText,
	)
	if isFav {
		text += "\n\n💚 В избранном"
	}
	return text
}
