package integration

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/cache"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
	"github.com/wywyy3cee/tgbot-anime-tracker/pkg/logger"
	"github.com/wywyy3cee/tgbot-anime-tracker/tests"
)

type RedisIntegrationSuite struct {
	suite.Suite
	cache       *cache.Cache
	redisClient *redis.Client
	logger      *logger.Logger
	testConfig  *tests.TestConfig
}

func (suite *RedisIntegrationSuite) SetupSuite() {
	var err error

	suite.logger = logger.New()
	suite.testConfig = tests.NewTestConfig()

	suite.redisClient = redis.NewClient(&redis.Options{
		Addr: suite.testConfig.RedisHost + ":" + suite.testConfig.RedisPort,
		DB:   suite.testConfig.RedisDB,
	})

	err = suite.redisClient.Ping(context.Background()).Err()
	require.NoError(suite.T(), err, "Failed to connect to Redis")

	redisURL := suite.testConfig.GetRedisURL()
	suite.cache, err = cache.New(redisURL, suite.logger)
	require.NoError(suite.T(), err, "Failed to initialize cache")
}

func (suite *RedisIntegrationSuite) TearDownSuite() {
	if suite.redisClient != nil {
		suite.redisClient.Close()
	}
}

func (suite *RedisIntegrationSuite) TearDownTest() {
	suite.redisClient.FlushDB(context.Background())
}

func (suite *RedisIntegrationSuite) TestAnimeSearchCache() {
	query := "test_anime"
	animes := []models.Anime{
		{
			ID:       1,
			Name:     "Test Anime",
			Russian:  "Тестовое Аниме",
			Kind:     "TV",
			Score:    "8.5",
			Status:   "released",
			Episodes: 12,
		},
	}

	err := suite.cache.SetAnimeSearch(query, animes, 1*time.Hour)
	assert.NoError(suite.T(), err, "Failed to set anime search cache")

	cached, err := suite.cache.GetAnimeSearch(query)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cached)
	assert.Equal(suite.T(), 1, len(cached))
	assert.Equal(suite.T(), "Test Anime", cached[0].Name)
}

func (suite *RedisIntegrationSuite) TestAnimeSearchCacheExpiration() {
	query := "expiring_anime"
	animes := []models.Anime{
		{
			ID:   1,
			Name: "Expiring Anime",
		},
	}

	err := suite.cache.SetAnimeSearch(query, animes, 100*time.Millisecond)
	assert.NoError(suite.T(), err)

	cached, err := suite.cache.GetAnimeSearch(query)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cached)

	time.Sleep(150 * time.Millisecond)

	cached, err = suite.cache.GetAnimeSearch(query)
	assert.Nil(suite.T(), cached)
}

func (suite *RedisIntegrationSuite) TestAnimeDetailsCache() {
	animeID := 1
	anime := &models.Anime{
		ID:          animeID,
		Name:        "Detailed Anime",
		Russian:     "Аниме с деталями",
		Kind:        "TV",
		Score:       "8.7",
		Status:      "released",
		Episodes:    24,
		Description: "Full description",
		Image: models.AnimeImage{
			Original: "https://example.com/original.png",
			Preview:  "/preview.png",
		},
	}

	err := suite.cache.SetAnimeDetails(animeID, anime, 24*time.Hour)
	assert.NoError(suite.T(), err, "Failed to set anime details cache")

	cached, err := suite.cache.GetAnimeDetails(animeID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cached)
	assert.Equal(suite.T(), "Detailed Anime", cached.Name)
	assert.Equal(suite.T(), 24, cached.Episodes)
	assert.Equal(suite.T(), "Full description", cached.Description)
}

func (suite *RedisIntegrationSuite) TestCacheMultipleQueries() {
	queries := []string{"query1", "query2", "query3"}
	animes := []models.Anime{
		{ID: 1, Name: "Anime 1"},
		{ID: 2, Name: "Anime 2"},
		{ID: 3, Name: "Anime 3"},
	}

	for i, q := range queries {
		err := suite.cache.SetAnimeSearch(q, []models.Anime{animes[i]}, 1*time.Hour)
		assert.NoError(suite.T(), err)
	}

	for i, q := range queries {
		cached, err := suite.cache.GetAnimeSearch(q)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), cached)
		assert.Equal(suite.T(), 1, len(cached))
		assert.Equal(suite.T(), "Anime "+string(rune('1'+i)), cached[0].Name)
	}
}

func (suite *RedisIntegrationSuite) TestCacheEmptyResults() {
	query := "empty_query"
	emptyResults := []models.Anime{}

	err := suite.cache.SetAnimeSearch(query, emptyResults, 1*time.Hour)
	assert.NoError(suite.T(), err)

	cached, err := suite.cache.GetAnimeSearch(query)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cached)
	assert.Equal(suite.T(), 0, len(cached))
}

func (suite *RedisIntegrationSuite) TestCacheConcurrentAccess() {
	done := make(chan error, 10)

	for i := 0; i < 5; i++ {
		go func(index int) {
			query := "concurrent_" + string(rune('1'+index))
			animes := []models.Anime{{ID: index, Name: "Anime " + string(rune('1'+index))}}
			err := suite.cache.SetAnimeSearch(query, animes, 1*time.Hour)
			done <- err
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func(index int) {
			query := "concurrent_" + string(rune('1'+index))
			_, err := suite.cache.GetAnimeSearch(query)
			done <- err
		}(i)
	}

	for i := 0; i < 10; i++ {
		err := <-done
		assert.NoError(suite.T(), err, "Concurrent operation failed")
	}
}

func (suite *RedisIntegrationSuite) TestCacheDetailsMultipleAnimes() {
	animeIDs := []int{1, 2, 3, 4, 5}

	for i, id := range animeIDs {
		anime := &models.Anime{
			ID:       id,
			Name:     "Anime " + string(rune('1'+i)),
			Episodes: 12 + i*2,
		}
		err := suite.cache.SetAnimeDetails(id, anime, 24*time.Hour)
		assert.NoError(suite.T(), err)
	}

	for i, id := range animeIDs {
		cached, err := suite.cache.GetAnimeDetails(id)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), cached)
		assert.Equal(suite.T(), "Anime "+string(rune('1'+i)), cached.Name)
		assert.Equal(suite.T(), 12+i*2, cached.Episodes)
	}
}

func (suite *RedisIntegrationSuite) TestCacheOverwrite() {
	query := "overwrite_test"

	firstAnimes := []models.Anime{{ID: 1, Name: "First Anime"}}
	err := suite.cache.SetAnimeSearch(query, firstAnimes, 1*time.Hour)
	assert.NoError(suite.T(), err)

	cached, _ := suite.cache.GetAnimeSearch(query)
	assert.Equal(suite.T(), "First Anime", cached[0].Name)

	secondAnimes := []models.Anime{{ID: 2, Name: "Second Anime"}}
	err = suite.cache.SetAnimeSearch(query, secondAnimes, 1*time.Hour)
	assert.NoError(suite.T(), err)

	cached, _ = suite.cache.GetAnimeSearch(query)
	assert.Equal(suite.T(), "Second Anime", cached[0].Name)
}

func TestRedisIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RedisIntegrationSuite))
}
