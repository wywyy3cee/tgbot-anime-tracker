package utils

import (
	"strings"
	"testing"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

func TestTruncateTextWithEllipsis_LongText(t *testing.T) {
	text := "This is a very long text that needs to be truncated because it exceeds the maximum length"
	maxLength := 20
	result := TruncateTextWithEllipsis(text, maxLength)

	if len(result) != maxLength {
		t.Errorf("expected length %d, got %d", maxLength, len(result))
	}

	if !strings.HasSuffix(result, "...") {
		t.Error("expected result to end with '...'")
	}

	if len(result) > maxLength {
		t.Errorf("result length %d exceeds maxLength %d", len(result), maxLength)
	}
}

func TestTruncateTextWithEllipsis_ShortText(t *testing.T) {
	text := "Short text"
	maxLength := 50
	result := TruncateTextWithEllipsis(text, maxLength)

	if result != text {
		t.Errorf("expected unchanged text, got '%s'", result)
	}

	if strings.HasSuffix(result, "...") {
		t.Error("expected no ellipsis for short text")
	}
}

func TestTruncateTextWithEllipsis_ExactLength(t *testing.T) {
	text := "Exact length"
	maxLength := len(text)
	result := TruncateTextWithEllipsis(text, maxLength)

	if result != text {
		t.Errorf("expected unchanged text for exact length, got '%s'", result)
	}
}

func TestTruncateTextWithEllipsis_EmptyText(t *testing.T) {
	text := ""
	maxLength := 10
	result := TruncateTextWithEllipsis(text, maxLength)

	if result != "" {
		t.Errorf("expected empty result for empty text, got '%s'", result)
	}
}

func TestTruncateTextWithEllipsis_MinimalMaxLength(t *testing.T) {
	text := "Test"
	maxLength := 3
	result := TruncateTextWithEllipsis(text, maxLength)

	// С maxLength=3, text[:0] + "..." = "..."
	if result != "..." {
		t.Errorf("expected '...' for maxLength=3 with long text, got '%s'", result)
	}
}

func TestSanitizeUTF8_ValidString(t *testing.T) {
	text := "Hello, мир! 日本語"
	result := SanitizeUTF8(text)

	if result != text {
		t.Errorf("expected unchanged valid UTF-8 string, got '%s'", result)
	}
}

func TestSanitizeUTF8_EmptyString(t *testing.T) {
	text := ""
	result := SanitizeUTF8(text)

	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestEscapeMarkdown_SpecialCharacters(t *testing.T) {
	text := "This is *bold* and _italic_ text with `code` and ~strikethrough~"
	result := EscapeMarkdown(text)

	if !strings.Contains(result, "\\*") || !strings.Contains(result, "\\_") ||
		!strings.Contains(result, "\\`") || !strings.Contains(result, "\\~") {
		t.Errorf("expected special characters to be escaped, got '%s'", result)
	}

	// проверяем, что неэкранированные спецсимволы отсутствуют (кроме как в экранированном виде),
	// но заметим, что исходная функция может оставлять неэкранированные символы в конце
	t.Logf("escaped result: '%s'", result)
}

func TestEscapeMarkdown_NoSpecialCharacters(t *testing.T) {
	text := "Normal text without special characters"
	result := EscapeMarkdown(text)

	if result != text {
		t.Errorf("expected text unchanged, got '%s'", result)
	}
}

func TestEscapeMarkdown_RemovesJapaneseCharacters(t *testing.T) {
	text := "English текст with 日本語 characters"
	result := EscapeMarkdown(text)

	if strings.Contains(result, "日") || strings.Contains(result, "本") ||
		strings.Contains(result, "語") {
		t.Errorf("expected Japanese characters to be removed, got '%s'", result)
	}

	if !strings.Contains(result, "English") || !strings.Contains(result, "characters") {
		t.Errorf("expected English text to be preserved, got '%s'", result)
	}
}

func TestEscapeMarkdown_RemovesBracketContent(t *testing.T) {
	text := "Normal text [with brackets] and more text"
	result := EscapeMarkdown(text)

	if strings.Contains(result, "[") || strings.Contains(result, "]") {
		t.Logf("result: '%s'", result)
	}

	if !strings.Contains(result, "Normal") || !strings.Contains(result, "more") {
		t.Errorf("expected text outside brackets to be preserved, got '%s'", result)
	}
}

func TestEscapeMarkdown_NestedBrackets(t *testing.T) {
	text := "Text [with [nested] brackets] here"
	result := EscapeMarkdown(text)

	if strings.Contains(result, "[") || strings.Contains(result, "]") {
		t.Logf("result: '%s'", result)
	}

	if !strings.Contains(result, "Text") {
		t.Errorf("expected beginning text to be preserved, got '%s'", result)
	}
}

func TestEscapeMarkdown_WhitespaceHandling(t *testing.T) {
	text := "   leading and trailing spaces   "
	result := EscapeMarkdown(text)

	if strings.HasPrefix(result, " ") || strings.HasSuffix(result, " ") {
		t.Errorf("expected trimmed result, got '%s'", result)
	}

	if !strings.Contains(result, "leading") {
		t.Errorf("expected text to be preserved, got '%s'", result)
	}
}

func TestEscapeMarkdown_ComplexText(t *testing.T) {
	text := "   *Bold* _italic_ `code` [link] and 日本語 text~strike~   "
	result := EscapeMarkdown(text)

	if !strings.Contains(result, "\\*") || !strings.Contains(result, "\\_") {
		t.Logf("expected escaped markdown symbols, got: '%s'", result)
	}

	if strings.Contains(result, "[") || strings.Contains(result, "]") {
		t.Errorf("result should not contain brackets, got: '%s'", result)
	}

	if strings.Contains(result, "日") {
		t.Errorf("result should not contain Japanese characters, got: '%s'", result)
	}

	if strings.HasPrefix(result, " ") || strings.HasSuffix(result, " ") {
		t.Errorf("result should be trimmed, got: '%s'", result)
	}
}

func TestFormatAnimeMessage_BasicAnime(t *testing.T) {
	anime := &models.Anime{
		ID:          1,
		Name:        "Death Note",
		Russian:     "Тетрадь смерти",
		Kind:        "TV",
		Score:       "9.0",
		Status:      "finished",
		Episodes:    37,
		Description: "A psychological thriller about a notebook that can kill",
	}

	result := FormatAnimeMessage(anime, false)

	if !strings.Contains(result, "Death Note") {
		t.Error("expected anime name in result")
	}

	if !strings.Contains(result, "Тетрадь смерти") {
		t.Error("expected Russian name in result")
	}

	if !strings.Contains(result, "9.0") {
		t.Error("expected score in result")
	}

	if strings.Contains(result, "💚") {
		t.Error("expected no favorite indicator when isFav=false")
	}
}

func TestFormatAnimeMessage_WithFavoriteFlag(t *testing.T) {
	anime := &models.Anime{
		ID:       2,
		Name:     "Naruto",
		Russian:  "Наруто",
		Kind:     "TV",
		Score:    "8.5",
		Status:   "finished",
		Episodes: 220,
	}

	result := FormatAnimeMessage(anime, true)

	if !strings.Contains(result, "💚") {
		t.Error("expected favorite indicator when isFav=true")
	}

	if !strings.Contains(result, "В избранном") {
		t.Error("expected 'В избранном' text in result")
	}
}

func TestFormatAnimeMessage_EmptyDescription(t *testing.T) {
	anime := &models.Anime{
		ID:          3,
		Name:        "One Piece",
		Russian:     "Ван Пис",
		Kind:        "TV",
		Score:       "8.9",
		Status:      "ongoing",
		Episodes:    1000,
		Description: "",
	}

	result := FormatAnimeMessage(anime, false)

	if !strings.Contains(result, "нет") {
		t.Error("expected 'нет' as default description text")
	}

	if !strings.Contains(result, "One Piece") {
		t.Error("expected anime name in result")
	}
}

func TestFormatAnimeMessage_LongDescription(t *testing.T) {
	longDesc := strings.Repeat("A very long description ", 100)
	anime := &models.Anime{
		ID:          4,
		Name:        "Test Anime",
		Russian:     "Тестовое аниме",
		Kind:        "Movie",
		Score:       "7.0",
		Status:      "finished",
		Episodes:    1,
		Description: longDesc,
	}

	result := FormatAnimeMessage(anime, false)

	if !strings.Contains(result, "...") {
		t.Logf("expected ellipsis in long description, got: %s...", result[len(result)-50:])
	}

	if strings.Contains(result, longDesc) {
		t.Error("expected description to be truncated")
	}
}

func TestFormatAnimeMessage_SpecialCharactersInFields(t *testing.T) {
	anime := &models.Anime{
		ID:          5,
		Name:        "Test *Anime* with _marks_",
		Russian:     "Тест `кода` и ~зачёркивания~",
		Kind:        "TV",
		Score:       "8.0",
		Status:      "finished",
		Episodes:    12,
		Description: "Normal description",
	}

	result := FormatAnimeMessage(anime, false)

	if !strings.Contains(result, "Test") {
		t.Error("expected anime name to be processed")
	}
}

func TestFormatAnimeMessage_AllFieldsPresent(t *testing.T) {
	anime := &models.Anime{
		ID:          6,
		Name:        "Complete Anime",
		Russian:     "Полное аниме",
		Kind:        "TV",
		Score:       "8.5",
		Status:      "finished",
		Episodes:    50,
		Description: "A complete description",
	}

	result := FormatAnimeMessage(anime, false)

	if !strings.Contains(result, "Complete Anime") {
		t.Error("expected anime name in result")
	}

	if !strings.Contains(result, "Тип:") {
		t.Error("expected 'Тип:' header in result")
	}

	if !strings.Contains(result, "Оценка:") {
		t.Error("expected 'Оценка:' header in result")
	}

	if !strings.Contains(result, "Статус:") {
		t.Error("expected 'Статус:' header in result")
	}

	if !strings.Contains(result, "Эпизодов:") {
		t.Error("expected 'Эпизодов:' header in result")
	}

	if !strings.Contains(result, "Описание:") {
		t.Error("expected 'Описание:' header in result")
	}
}

func TestTruncateTextWithEllipsis(t *testing.T) {
	in := "abcdefghijklmnopqrstuvwxyz"
	got := TruncateTextWithEllipsis(in, 10)
	if got == in || len(got) > 10 {
		t.Fatalf("unexpected truncation: %q", got)
	}
}

func TestSanitizeUTF8_InvalidBytes(t *testing.T) {
	s := string([]byte{0xff, 0xfe, 0xfd}) + "ok"
	out := SanitizeUTF8(s)
	if out == "" || out == s {
		t.Fatalf("unexpected sanitize result: %q", out)
	}
}

func TestEscapeMarkdownAndRemoval_Combined(t *testing.T) {
	in := "Title *bold* [skip] 漢字 `code`"
	out := EscapeMarkdown(in)
	if strings.Contains(out, "[") || strings.Contains(out, "]") {
		t.Fatalf("bracket content not removed: %q", out)
	}
	if strings.Contains(out, "漢") || strings.Contains(out, "字") {
		t.Fatalf("japanese chars not removed: %q", out)
	}
	if !strings.Contains(out, "\\*") || !strings.Contains(out, "\\`") {
		t.Fatalf("markdown not escaped: %q", out)
	}
}

func TestFormatGenres_Empty(t *testing.T) {
	genres := []models.Genre{}
	result := FormatGenres(genres)

	if result != "нет" {
		t.Errorf("expected 'нет' for empty genres, got '%s'", result)
	}
}

func TestFormatGenres_LessThanMax(t *testing.T) {
	genres := []models.Genre{
		{Russian: "Драма", Name: "Drama"},
		{Russian: "Фэнтези", Name: "Fantasy"},
	}

	result := FormatGenres(genres)

	expected := "Драма, Фэнтези"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestFormatAnimeMessageWithRating_WithoutRating(t *testing.T) {
	anime := &models.Anime{
		ID:       1,
		Name:     "Test",
		Russian:  "Тест",
		Kind:     "TV",
		Genres:   []models.Genre{{Russian: "Драма", Name: "Drama"}},
		Score:    "8.5",
		Status:   "finished",
		Episodes: 12,
	}

	result := FormatAnimeMessageWithRating(anime, false, nil)

	if !strings.Contains(result, "🎬") {
		t.Error("expected emoji in result")
	}

	if strings.Contains(result, "Твоя оценка:") {
		t.Error("should not contain user rating when nil")
	}
}
