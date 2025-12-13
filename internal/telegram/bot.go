package telegram

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/shikimori"
)

type Bot struct {
	api             *tgbotapi.BotAPI
	shikimoriClient *shikimori.Client

	userStates map[int64]*UserState
	mu         sync.RWMutex
}

type UserState struct {
	SearchResults []models.Anime
	CurrentIndex  int
}

func NewBot(token string, client *shikimori.Client) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		api:             api,
		shikimoriClient: client,
		userStates:      make(map[int64]*UserState),
	}, nil
}

func (b *Bot) saveState(userID int64, state *UserState) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.userStates[userID] = state
}

func (b *Bot) getState(userID int64) *UserState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.userStates[userID]
}

func (b *Bot) getCurrentAnime(userID int64) *models.Anime {
	state := b.getState(userID)
	if state == nil {
		return nil
	}

	if state.CurrentIndex >= len(state.SearchResults) {
		return nil
	}

	return &state.SearchResults[state.CurrentIndex]
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		}
	}

	return nil
}
