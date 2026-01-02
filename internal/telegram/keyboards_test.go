package telegram

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

func TestCreateAnimeKeyboard_NoState(t *testing.T) {
	b := &Bot{userStates: make(map[int64]*UserState)}
	kb := b.createAnimeKeyboard(1, 10, false, nil)
	if len(kb.InlineKeyboard) != 0 {
		t.Fatalf("expected empty keyboard on no state, got: %v", kb)
	}
}

func TestCreateAnimeKeyboard_WithState(t *testing.T) {
	b := &Bot{userStates: make(map[int64]*UserState)}
	b.userStates[1] = &UserState{SearchResults: []models.Anime{{ID: 1}, {ID: 2}}, CurrentIndex: 0}
	kb := b.createAnimeKeyboard(1, 1, true, nil)
	if len(kb.InlineKeyboard) == 0 {
		t.Fatalf("expected keyboard rows, got none")
	}
}

func TestCreateFavoritesKeyboard_Pagination(t *testing.T) {
	b := &Bot{}
	favs := []models.Favorite{}
	for i := 0; i < 25; i++ {
		favs = append(favs, models.Favorite{AnimeID: i, Title: "title"})
	}
	kb := b.createFavoritesKeyboard(favs, 1, 3)
	if len(kb.InlineKeyboard) == 0 {
		t.Fatalf("expected favorites keyboard rows, got none")
	}
}

func TestCreateMainAndCancelKeyboards(t *testing.T) {
	b := &Bot{}
	m := b.createMainMenuKeyboard()
	if len(m.Keyboard) == 0 {
		t.Fatalf("main menu keyboard empty")
	}
	c := b.createCancelKeyboard()
	if len(c.Keyboard) == 0 {
		t.Fatalf("cancel keyboard empty")
	}
	_ = tgbotapi.NewInlineKeyboardMarkup()
}

func TestCreateAnimeKeyboard_WithRating(t *testing.T) {
	b := &Bot{userStates: make(map[int64]*UserState)}
	b.userStates[1] = &UserState{
		SearchResults: []models.Anime{{ID: 1}, {ID: 2}},
		CurrentIndex:  0,
	}
	rating := &models.Rating{Score: 8}
	kb := b.createAnimeKeyboard(1, 1, true, rating)
	if len(kb.InlineKeyboard) == 0 {
		t.Fatalf("expected keyboard rows, got none")
	}
}

func TestCreateAnimeKeyboard_MultipleResults(t *testing.T) {
	b := &Bot{userStates: make(map[int64]*UserState)}
	animes := []models.Anime{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}}
	b.userStates[1] = &UserState{
		SearchResults: animes,
		CurrentIndex:  2,
	}
	kb := b.createAnimeKeyboard(1, 1, false, nil)
	if len(kb.InlineKeyboard) < 2 {
		t.Error("expected navigation and action rows")
	}
}

func TestCreateAnimeKeyboard_FirstPosition(t *testing.T) {
	b := &Bot{userStates: make(map[int64]*UserState)}
	animes := []models.Anime{{ID: 1}, {ID: 2}}
	b.userStates[1] = &UserState{
		SearchResults: animes,
		CurrentIndex:  0,
	}
	kb := b.createAnimeKeyboard(1, 1, false, nil)
	if len(kb.InlineKeyboard) == 0 {
		t.Error("expected keyboard rows")
	}
}

func TestCreateAnimeKeyboard_LastPosition(t *testing.T) {
	b := &Bot{userStates: make(map[int64]*UserState)}
	animes := []models.Anime{{ID: 1}, {ID: 2}}
	b.userStates[1] = &UserState{
		SearchResults: animes,
		CurrentIndex:  1,
	}
	kb := b.createAnimeKeyboard(1, 2, false, nil)
	if len(kb.InlineKeyboard) == 0 {
		t.Error("expected keyboard rows")
	}
}

func TestCreateRatingKeyboard(t *testing.T) {
	b := &Bot{}
	kb := b.createRatingKeyboard(1)
	if len(kb.InlineKeyboard) != 3 {
		t.Errorf("expected 3 rows (1-5, 6-10, cancel), got %d", len(kb.InlineKeyboard))
	}
	if len(kb.InlineKeyboard[0]) != 5 {
		t.Errorf("expected 5 buttons in first row, got %d", len(kb.InlineKeyboard[0]))
	}
	if len(kb.InlineKeyboard[1]) != 5 {
		t.Errorf("expected 5 buttons in second row, got %d", len(kb.InlineKeyboard[1]))
	}
	if len(kb.InlineKeyboard[2]) != 1 {
		t.Errorf("expected 1 cancel button, got %d", len(kb.InlineKeyboard[2]))
	}
}

func TestCreateFavoritesKeyboard_FirstPage(t *testing.T) {
	b := &Bot{}
	favs := []models.Favorite{
		{ID: 1, Title: "Anime 1"},
		{ID: 2, Title: "Anime 2"},
	}
	kb := b.createFavoritesKeyboard(favs, 0, 1)
	if len(kb.InlineKeyboard) < 2 {
		t.Error("expected at least 2 rows (items + navigation)")
	}
}

func TestCreateFavoritesKeyboard_MiddlePage(t *testing.T) {
	b := &Bot{}
	var favs []models.Favorite
	for i := 0; i < 35; i++ {
		favs = append(favs, models.Favorite{ID: i + 1, Title: "Anime " + string(rune(i))})
	}
	kb := b.createFavoritesKeyboard(favs, 1, 4)
	if len(kb.InlineKeyboard) == 0 {
		t.Error("expected keyboard rows")
	}
}

func TestCreateFavoritesKeyboard_LastPage(t *testing.T) {
	b := &Bot{}
	var favs []models.Favorite
	for i := 0; i < 15; i++ {
		favs = append(favs, models.Favorite{ID: i + 1, Title: "Anime " + string(rune(i))})
	}
	kb := b.createFavoritesKeyboard(favs, 1, 2)
	if len(kb.InlineKeyboard) == 0 {
		t.Error("expected keyboard rows")
	}
}

func TestCreateFavoritesKeyboard_LongTitles(t *testing.T) {
	b := &Bot{}
	longTitle := "This is a very long anime title that should be truncated to fit in the button"
	favs := []models.Favorite{
		{ID: 1, Title: longTitle},
	}
	kb := b.createFavoritesKeyboard(favs, 0, 1)
	if len(kb.InlineKeyboard) == 0 {
		t.Error("expected keyboard rows")
	}
	if len(kb.InlineKeyboard[0]) > 0 {
		buttonText := kb.InlineKeyboard[0][0].Text
		if len(buttonText) > 64 {
			t.Errorf("button text is too long: %d characters", len(buttonText))
		}
	}
}

func TestCreateFavoritesKeyboard_EmptyList(t *testing.T) {
	b := &Bot{}
	favs := []models.Favorite{}
	kb := b.createFavoritesKeyboard(favs, 0, 0)
	if len(kb.InlineKeyboard) > 1 {
		t.Errorf("expected at most 1 row for empty list, got %d", len(kb.InlineKeyboard))
	}
}

func TestCreateFavoritesKeyboard_SingleItem(t *testing.T) {
	b := &Bot{}
	favs := []models.Favorite{{ID: 1, Title: "Single Anime"}}
	kb := b.createFavoritesKeyboard(favs, 0, 1)
	if len(kb.InlineKeyboard) == 0 {
		t.Error("expected keyboard rows")
	}
}

func TestCreateFavoriteAnimeKeyboard(t *testing.T) {
	b := &Bot{}
	kb := b.createFavoriteAnimeKeyboard(1, nil)
	if len(kb.InlineKeyboard) < 3 {
		t.Errorf("expected at least 3 rows, got %d", len(kb.InlineKeyboard))
	}
}

func TestCreateFavoriteAnimeKeyboard_WithRatingBasic(t *testing.T) {
	b := &Bot{}
	rating := &models.Rating{Score: 7}
	kb := b.createFavoriteAnimeKeyboard(1, rating)
	if len(kb.InlineKeyboard) < 3 {
		t.Error("expected at least 3 rows")
	}
}

func TestCreateCancelKeyboard(t *testing.T) {
	b := &Bot{}
	kb := b.createCancelKeyboard()
	if len(kb.Keyboard) == 0 {
		t.Error("expected reply keyboard rows")
	}
	if len(kb.Keyboard[0]) != 1 {
		t.Errorf("expected 1 button in cancel keyboard, got %d", len(kb.Keyboard[0]))
	}
}

func TestCreateMainMenuKeyboard_Structure(t *testing.T) {
	b := &Bot{}
	kb := b.createMainMenuKeyboard()
	if len(kb.Keyboard) < 1 {
		t.Error("expected at least 1 row in main menu")
	}
	if len(kb.Keyboard[0]) != 2 {
		t.Errorf("expected 2 buttons in first row, got %d", len(kb.Keyboard[0]))
	}
	if len(kb.Keyboard[1]) != 1 {
		t.Errorf("expected 1 button in second row, got %d", len(kb.Keyboard[1]))
	}
}
