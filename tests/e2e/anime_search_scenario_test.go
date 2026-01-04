package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/cache"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/database"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/service"
	"github.com/wywyy3cee/tgbot-anime-tracker/pkg/logger"
	"github.com/wywyy3cee/tgbot-anime-tracker/tests"
	"github.com/wywyy3cee/tgbot-anime-tracker/tests/mocks"
)

type AnimeSearchScenarioSuite struct {
	suite.Suite
	db           *database.Database
	cache        *cache.Cache
	animeService *service.AnimeService
	repository   *database.Repository
	mockClient   *mocks.MockShikimoriClient
	logger       *logger.Logger
	redisClient  *redis.Client
	testConfig   *tests.TestConfig
}

func (suite *AnimeSearchScenarioSuite) SetupSuite() {
	var err error

	suite.logger = logger.New()
	suite.testConfig = tests.NewTestConfig()

	dbURL := suite.testConfig.GetDatabaseURL()
	suite.db, err = database.Connect(dbURL, suite.logger)
	require.NoError(suite.T(), err, "Failed to connect to database")

	err = suite.db.RunMigrations(suite.testConfig.MigrationsDir)
	require.NoError(suite.T(), err, "Failed to run migrations")

	suite.redisClient = redis.NewClient(&redis.Options{
		Addr: suite.testConfig.RedisHost + ":" + suite.testConfig.RedisPort,
		DB:   suite.testConfig.RedisDB,
	})
	err = suite.redisClient.Ping(context.Background()).Err()
	require.NoError(suite.T(), err, "Failed to connect to Redis")

	redisURL := suite.testConfig.GetRedisURL()
	suite.cache, err = cache.New(redisURL, suite.logger)
	require.NoError(suite.T(), err, "Failed to initialize cache")

	suite.redisClient.FlushDB(context.Background())

	suite.repository = database.NewRepository(suite.db)
	suite.mockClient = mocks.NewMockShikimoriClient()

	suite.animeService = service.NewAnimeService(nil, suite.repository, suite.cache)
	suite.animeService.SetShikimoriClient(suite.mockClient)
}

func (suite *AnimeSearchScenarioSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.redisClient != nil {
		suite.redisClient.Close()
	}
}

func (suite *AnimeSearchScenarioSuite) TearDownTest() {
	suite.redisClient.FlushDB(context.Background())
	suite.db.DB.Exec("TRUNCATE TABLE users CASCADE")
	suite.db.DB.Exec("TRUNCATE TABLE favorites CASCADE")
	suite.db.DB.Exec("TRUNCATE TABLE ratings CASCADE")
}

// полный e2e сценарий поиска аниме
func (suite *AnimeSearchScenarioSuite) TestCompleteAnimeSearchUserStory() {
	userID := int64(12345)
	username := "testuser"

	suite.Run("Step1_UserInitiatesSearch", func() {
		err := suite.animeService.EnsureUserExists(userID, username)
		assert.NoError(suite.T(), err, "Failed to create user")

		user, err := suite.repository.GetUser(userID)
		assert.NoError(suite.T(), err, "Failed to get user")
		assert.NotNil(suite.T(), user)
		assert.Equal(suite.T(), username, user.Username)
	})

	suite.Run("Step2_UserSearchesAnime", func() {
		mockAnimes := suite.getMockAnimes()
		suite.mockClient.SetSearchResults("Naruto", mockAnimes)

		results, err := suite.animeService.SearchAnime("Naruto")
		assert.NoError(suite.T(), err, "Failed to search anime")
		assert.NotEmpty(suite.T(), results, "No results returned")
		assert.Greater(suite.T(), len(results), 0, "Expected results but got none")

		for _, anime := range results {
			assert.NotEmpty(suite.T(), anime.Name, "Anime name is empty")
			assert.NotEmpty(suite.T(), anime.Score, "Anime score is empty")
			assert.NotEmpty(suite.T(), anime.Status, "Anime status is empty")
			assert.Greater(suite.T(), anime.Episodes, 0, "Episode count is 0")
			assert.NotEmpty(suite.T(), anime.Genres, "No genres provided")
			assert.NotEmpty(suite.T(), anime.Image.Original, "No poster URL")
		}
	})

	suite.Run("Step3_VerifyCacheFunctionality", func() {
		mockAnimes := suite.getMockAnimes()
		query := "Demon Slayer"
		suite.mockClient.SetSearchResults(query, mockAnimes)
		suite.mockClient.ResetCallCount()

		results1, err := suite.animeService.SearchAnime(query)
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), results1)

		callCountAfterFirst := suite.mockClient.GetSearchCallCount()

		results2, err := suite.animeService.SearchAnime(query)
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), results2)

		callCountAfterSecond := suite.mockClient.GetSearchCallCount()
		assert.Equal(suite.T(), callCountAfterFirst, callCountAfterSecond, "Cache should prevent API call")
		assert.Equal(suite.T(), results1, results2, "Cached results differ from original")
	})

	suite.Run("Step4_UserAddsAnimeToFavorites", func() {
		mockAnimes := suite.getMockAnimes()
		selectedAnime := mockAnimes[0]

		err := suite.animeService.AddToFavorites(userID, selectedAnime)
		assert.NoError(suite.T(), err, "Failed to add anime to favorites")

		isFavorite, err := suite.animeService.IsFavorite(userID, selectedAnime.ID)
		assert.NoError(suite.T(), err, "Failed to check if favorite")
		assert.True(suite.T(), isFavorite, "Anime should be marked as favorite")

		favorites, err := suite.animeService.GetUserFavorites(userID)
		assert.NoError(suite.T(), err, "Failed to get favorites")
		assert.NotEmpty(suite.T(), favorites, "Favorites list is empty")
		assert.Equal(suite.T(), 1, len(favorites), "Expected 1 favorite")
		assert.Equal(suite.T(), selectedAnime.ID, favorites[0].AnimeID)
	})

	suite.Run("Step5_UserRatesAnime", func() {
		mockAnimes := suite.getMockAnimes()
		selectedAnime := mockAnimes[0]

		rating := 8
		err := suite.animeService.AddRating(userID, selectedAnime.ID, rating)
		assert.NoError(suite.T(), err, "Failed to add rating")

		userRating, err := suite.animeService.GetUserRating(userID, selectedAnime.ID)
		assert.NoError(suite.T(), err, "Failed to get rating")
		assert.NotNil(suite.T(), userRating)
		assert.Equal(suite.T(), rating, userRating.Score)
	})

	suite.Run("Step6_UserRemovesFromFavorites", func() {
		mockAnimes := suite.getMockAnimes()
		selectedAnime := mockAnimes[0]

		err := suite.animeService.RemoveFromFavorites(userID, selectedAnime.ID)
		assert.NoError(suite.T(), err, "Failed to remove from favorites")

		isFavorite, err := suite.animeService.IsFavorite(userID, selectedAnime.ID)
		assert.NoError(suite.T(), err)
		assert.False(suite.T(), isFavorite, "Anime should not be in favorites")

		favorites, err := suite.animeService.GetUserFavorites(userID)
		assert.NoError(suite.T(), err)
		assert.Empty(suite.T(), favorites, "Favorites should be empty")
	})

	suite.Run("Step7_SearchWithNoResults", func() {
		suite.mockClient.SetSearchResults("NonExistentAnime12345", []models.Anime{})

		_, err := suite.animeService.SearchAnime("NonExistentAnime12345")
		assert.Error(suite.T(), err, "Expected error for empty search results")
	})

	suite.Run("Step8_MultipleFavoritesAndRatings", func() {
		mockAnimes := suite.getMockAnimes()

		for i := 0; i < 3 && i < len(mockAnimes); i++ {
			err := suite.animeService.AddToFavorites(userID, mockAnimes[i])
			assert.NoError(suite.T(), err)

			err = suite.animeService.AddRating(userID, mockAnimes[i].ID, 7+i)
			assert.NoError(suite.T(), err)
		}

		favorites, err := suite.animeService.GetUserFavorites(userID)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 3, len(favorites), "Expected 3 favorites")

		for i := 0; i < 3 && i < len(mockAnimes); i++ {
			rating, err := suite.animeService.GetUserRating(userID, mockAnimes[i].ID)
			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), rating)
			assert.Equal(suite.T(), 7+i, rating.Score)
		}
	})
}

func (suite *AnimeSearchScenarioSuite) TestSearchPerformance() {
	mockAnimes := suite.getMockAnimes()
	suite.mockClient.SetSearchResults("Performance", mockAnimes)

	start := time.Now()
	_, err := suite.animeService.SearchAnime("Performance")
	duration := time.Since(start)

	assert.NoError(suite.T(), err)
	assert.Less(suite.T(), duration, 500*time.Millisecond, "Search took too long")
}

func (suite *AnimeSearchScenarioSuite) TestCacheEffectiveness() {
	mockAnimes := suite.getMockAnimes()
	query := "CacheTest"
	suite.mockClient.SetSearchResults(query, mockAnimes)

	start1 := time.Now()
	_, err := suite.animeService.SearchAnime(query)
	duration1 := time.Since(start1)
	assert.NoError(suite.T(), err)

	start2 := time.Now()
	_, err = suite.animeService.SearchAnime(query)
	duration2 := time.Since(start2)
	assert.NoError(suite.T(), err)

	assert.Less(suite.T(), duration2, duration1, "Cached result should be faster")
}

func (suite *AnimeSearchScenarioSuite) TestConcurrentSearches() {
	mockAnimes := suite.getMockAnimes()
	queries := []string{"Concurrent1", "Concurrent2", "Concurrent3"}

	for _, q := range queries {
		suite.mockClient.SetSearchResults(q, mockAnimes)
	}

	results := make(chan error, len(queries))

	for _, q := range queries {
		go func(query string) {
			_, err := suite.animeService.SearchAnime(query)
			results <- err
		}(q)
	}

	for i := 0; i < len(queries); i++ {
		err := <-results
		assert.NoError(suite.T(), err, "Concurrent search failed")
	}
}

func (suite *AnimeSearchScenarioSuite) getMockAnimes() []models.Anime {
	return []models.Anime{
		{
			ID:          1,
			Name:        "Naruto",
			Russian:     "Наруто",
			Kind:        "TV",
			Score:       "8.5",
			Status:      "released",
			Episodes:    220,
			AiredOn:     "2002-10-03",
			ReleasedOn:  "2007-02-08",
			Description: "Naruto Uzumaki is a young shinobi with an indomitable ninja spirit.",
			Genres: []models.Genre{
				{ID: 1, Name: "action", Russian: "Экшен", Kind: "anime", EntryType: "genre"},
				{ID: 2, Name: "adventure", Russian: "Приключения", Kind: "anime", EntryType: "genre"},
				{ID: 3, Name: "super power", Russian: "Суперсила", Kind: "anime", EntryType: "genre"},
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
			AiredOn:     "2019-04-06",
			ReleasedOn:  "2019-09-28",
			Description: "Tanjiro sets out on the path of the Demon Slayer to save his sister.",
			Genres: []models.Genre{
				{ID: 1, Name: "action", Russian: "Экшен", Kind: "anime", EntryType: "genre"},
				{ID: 2, Name: "adventure", Russian: "Приключения", Kind: "anime", EntryType: "genre"},
				{ID: 4, Name: "shounen", Russian: "Сёнэн", Kind: "anime", EntryType: "genre"},
			},
			Image: models.AnimeImage{
				Original: "https://shikimori.one/system/animes/original/2.png",
				Preview:  "/system/animes/preview/2.png",
			},
		},
		{
			ID:          3,
			Name:        "Attack on Titan",
			Russian:     "Атака титанов",
			Kind:        "TV",
			Score:       "8.9",
			Status:      "released",
			Episodes:    86,
			AiredOn:     "2013-04-07",
			ReleasedOn:  "2023-11-04",
			Description: "Humanity's fight against the colossal Titans.",
			Genres: []models.Genre{
				{ID: 1, Name: "action", Russian: "Экшен", Kind: "anime", EntryType: "genre"},
				{ID: 2, Name: "adventure", Russian: "Приключения", Kind: "anime", EntryType: "genre"},
				{ID: 5, Name: "psychological", Russian: "Психологический", Kind: "anime", EntryType: "genre"},
			},
			Image: models.AnimeImage{
				Original: "https://shikimori.one/system/animes/original/3.png",
				Preview:  "/system/animes/preview/3.png",
			},
		},
	}
}

func TestAnimeSearchScenarioSuite(t *testing.T) {
	suite.Run(t, new(AnimeSearchScenarioSuite))
}
