package integration

import (
	"context"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/cache"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/database"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/service"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/telegram"
	"github.com/wywyy3cee/tgbot-anime-tracker/pkg/logger"
	"github.com/wywyy3cee/tgbot-anime-tracker/tests"
	"github.com/wywyy3cee/tgbot-anime-tracker/tests/mocks"
)

type TelegramHandlersIntegrationSuite struct {
	suite.Suite
	db           *database.Database
	cache        *cache.Cache
	repository   *database.Repository
	animeService *service.AnimeService
	mockClient   *mocks.MockShikimoriClient
	mockBotAPI   *mocks.MockTelegramBot
	bot          *telegram.Bot
	logger       *logger.Logger
	redisClient  *redis.Client
	testConfig   *tests.TestConfig
}

func (suite *TelegramHandlersIntegrationSuite) SetupSuite() {
	var err error

	suite.logger = logger.New()
	suite.testConfig = tests.NewTestConfig()

	dbURL := suite.testConfig.GetDatabaseURL()
	suite.db, err = database.Connect(dbURL, suite.logger)
	require.NoError(suite.T(), err)

	err = suite.db.RunMigrations(suite.testConfig.MigrationsDir)
	require.NoError(suite.T(), err)

	suite.redisClient = redis.NewClient(&redis.Options{
		Addr: suite.testConfig.RedisHost + ":" + suite.testConfig.RedisPort,
		DB:   suite.testConfig.RedisDB,
	})
	err = suite.redisClient.Ping(context.Background()).Err()
	require.NoError(suite.T(), err)

	redisURL := suite.testConfig.GetRedisURL()
	suite.cache, err = cache.New(redisURL, suite.logger)
	require.NoError(suite.T(), err)

	suite.redisClient.FlushDB(context.Background())

	suite.repository = database.NewRepository(suite.db)
	suite.mockClient = mocks.NewMockShikimoriClient()

	suite.animeService = service.NewAnimeService(nil, suite.repository, suite.cache)
	suite.animeService.SetShikimoriClient(suite.mockClient)

	suite.mockBotAPI = mocks.NewMockTelegramBot()
	suite.bot, err = telegram.NewBotWithAPI(suite.mockBotAPI.API, suite.animeService, suite.logger)
	require.NoError(suite.T(), err)
}

func (suite *TelegramHandlersIntegrationSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.redisClient != nil {
		suite.redisClient.Close()
	}
}

func (suite *TelegramHandlersIntegrationSuite) TearDownTest() {
	suite.redisClient.FlushDB(context.Background())
	suite.db.DB.Exec("TRUNCATE TABLE users CASCADE")
	suite.db.DB.Exec("TRUNCATE TABLE favorites CASCADE")
	suite.db.DB.Exec("TRUNCATE TABLE ratings CASCADE")
	suite.mockClient.ResetCallCount()
	suite.mockBotAPI.Reset()
}

func (suite *TelegramHandlersIntegrationSuite) TestUserSearchFlow() {
	userID := int64(12345)
	username := "testuser"

	mockAnimes := suite.getMockAnimes()
	suite.mockClient.SetSearchResults("naruto", mockAnimes)

	suite.animeService.EnsureUserExists(userID, username)

	user, err := suite.repository.GetUser(userID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), username, user.Username)

	results, err := suite.animeService.SearchAnime("naruto")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(results))
	assert.Equal(suite.T(), "Naruto", results[0].Name)
	assert.Equal(suite.T(), "Demon Slayer", results[1].Name)

	assert.Equal(suite.T(), 1, suite.mockClient.GetSearchCallCount())

	results2, err := suite.animeService.SearchAnime("naruto")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(results2))
	assert.Equal(suite.T(), 1, suite.mockClient.GetSearchCallCount())
}

func (suite *TelegramHandlersIntegrationSuite) TestFavoritesFlow() {
	userID := int64(12345)
	chatID := int64(67890)
	username := "testuser"

	suite.animeService.EnsureUserExists(userID, username)
	mockAnimes := suite.getMockAnimes()
	suite.mockClient.SetSearchResults("test", mockAnimes)

	searchMsg := &tgbotapi.Message{
		From: &tgbotapi.User{ID: userID, UserName: username, FirstName: "Test"},
		Chat: &tgbotapi.Chat{ID: chatID},
		Text: "test",
	}

	suite.bot.HandleUpdate(&tgbotapi.Update{
		UpdateID: 1,
		Message:  searchMsg,
	})

	anime := mockAnimes[0]
	err := suite.animeService.AddToFavorites(userID, anime)
	assert.NoError(suite.T(), err)

	isFavorite, err := suite.animeService.IsFavorite(userID, anime.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isFavorite)

	favorites, err := suite.animeService.GetUserFavorites(userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(favorites))
	assert.Equal(suite.T(), anime.ID, favorites[0].AnimeID)
}

func (suite *TelegramHandlersIntegrationSuite) TestRatingFlow() {
	userID := int64(12345)
	username := "testuser"

	suite.animeService.EnsureUserExists(userID, username)
	mockAnimes := suite.getMockAnimes()
	animeID := mockAnimes[0].ID

	err := suite.animeService.AddRating(userID, animeID, 8)
	assert.NoError(suite.T(), err)

	rating, err := suite.animeService.GetUserRating(userID, animeID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), rating)
	assert.Equal(suite.T(), 8, rating.Score)

	err = suite.animeService.AddRating(userID, animeID, 9)
	assert.NoError(suite.T(), err)

	rating, err = suite.animeService.GetUserRating(userID, animeID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 9, rating.Score)
}

func (suite *TelegramHandlersIntegrationSuite) TestMultipleUsersIndependence() {
	user1ID := int64(111)
	user2ID := int64(222)
	username1 := "user1"
	username2 := "user2"

	suite.animeService.EnsureUserExists(user1ID, username1)
	suite.animeService.EnsureUserExists(user2ID, username2)

	mockAnimes := suite.getMockAnimes()

	err := suite.animeService.AddToFavorites(user1ID, mockAnimes[0])
	assert.NoError(suite.T(), err)

	err = suite.animeService.AddToFavorites(user2ID, mockAnimes[1])
	assert.NoError(suite.T(), err)

	isFav1, _ := suite.animeService.IsFavorite(user1ID, mockAnimes[0].ID)
	isFav2, _ := suite.animeService.IsFavorite(user1ID, mockAnimes[1].ID)
	assert.True(suite.T(), isFav1)
	assert.False(suite.T(), isFav2)

	isFav1, _ = suite.animeService.IsFavorite(user2ID, mockAnimes[0].ID)
	isFav2, _ = suite.animeService.IsFavorite(user2ID, mockAnimes[1].ID)
	assert.False(suite.T(), isFav1)
	assert.True(suite.T(), isFav2)
}

func (suite *TelegramHandlersIntegrationSuite) TestSearchWithNoResults() {
	suite.mockClient.SetSearchResults("nonexistent", []models.Anime{})

	_, err := suite.animeService.SearchAnime("nonexistent")
	assert.Error(suite.T(), err)
}

func (suite *TelegramHandlersIntegrationSuite) TestCacheEffectInMultipleSearches() {
	userID := int64(12345)
	username := "testuser"

	suite.animeService.EnsureUserExists(userID, username)

	mockAnimes := suite.getMockAnimes()
	query := "cached_search"
	suite.mockClient.SetSearchResults(query, mockAnimes)

	results1, err := suite.animeService.SearchAnime(query)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), results1)

	callCountAfterFirst := suite.mockClient.GetSearchCallCount()

	results2, err := suite.animeService.SearchAnime(query)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), results1, results2)

	callCountAfterSecond := suite.mockClient.GetSearchCallCount()
	assert.Equal(suite.T(), callCountAfterFirst, callCountAfterSecond)
}

func (suite *TelegramHandlersIntegrationSuite) getMockAnimes() []models.Anime {
	return []models.Anime{
		{
			ID:          1,
			Name:        "Naruto",
			Russian:     "Наруто",
			Kind:        "TV",
			Score:       "8.5",
			Status:      "released",
			Episodes:    220,
			Description: "A young ninja with a powerful spirit",
			Genres: []models.Genre{
				{ID: 1, Name: "action", Russian: "Экшен"},
				{ID: 2, Name: "adventure", Russian: "Приключения"},
			},
			Image: models.AnimeImage{
				Original: "https://shikimori.one/system/animes/original/1.png",
				Preview:  "/system/animes/preview/1.png",
			},
		},
		{
			ID:          2,
			Name:        "Demon Slayer",
			Russian:     "Истребитель демонов",
			Kind:        "TV",
			Score:       "8.7",
			Status:      "released",
			Episodes:    26,
			Description: "A journey to save a sister",
			Genres: []models.Genre{
				{ID: 1, Name: "action", Russian: "Экшен"},
			},
			Image: models.AnimeImage{
				Original: "https://shikimori.one/system/animes/original/2.png",
				Preview:  "/system/animes/preview/2.png",
			},
		},
	}
}

func TestTelegramHandlersIntegrationSuite(t *testing.T) {
	suite.Run(t, new(TelegramHandlersIntegrationSuite))
}
