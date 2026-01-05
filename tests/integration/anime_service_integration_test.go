package integration

import (
	"context"
	"testing"

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

type AnimeServiceIntegrationSuite struct {
	suite.Suite
	db            *database.Database
	cache         *cache.Cache
	repository    *database.Repository
	animeService  *service.AnimeService
	mockClient    *mocks.MockShikimoriClient
	logger        *logger.Logger
	redisClient   *redis.Client
	testConfig    *tests.TestConfig
}

func (suite *AnimeServiceIntegrationSuite) SetupSuite() {
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
}

func (suite *AnimeServiceIntegrationSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.redisClient != nil {
		suite.redisClient.Close()
	}
}

func (suite *AnimeServiceIntegrationSuite) TearDownTest() {
	suite.redisClient.FlushDB(context.Background())
	suite.db.DB.Exec("TRUNCATE TABLE users CASCADE")
	suite.db.DB.Exec("TRUNCATE TABLE favorites CASCADE")
	suite.db.DB.Exec("TRUNCATE TABLE ratings CASCADE")
	suite.mockClient.ResetCallCount()
}

func (suite *AnimeServiceIntegrationSuite) TestSearchAnimeWithCache() {
	mockAnimes := suite.getMockAnimes()
	query := "test_search"
	suite.mockClient.SetSearchResults(query, mockAnimes)

	results, err := suite.animeService.SearchAnime(query)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), len(mockAnimes), len(results))
	initialCallCount := suite.mockClient.GetSearchCallCount()

	results2, err := suite.animeService.SearchAnime(query)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), results, results2)
	assert.Equal(suite.T(), initialCallCount, suite.mockClient.GetSearchCallCount())
}

func (suite *AnimeServiceIntegrationSuite) TestGetAnimeByIDWithCache() {
	anime := &models.Anime{
		ID:      1,
		Name:    "Test Anime",
		Russian: "Тестовое Аниме",
		Score:   "8.5",
		Status:  "released",
	}
	suite.mockClient.SetAnimeDetails(1, anime)

	result, err := suite.animeService.GetAnimeByID(1)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Test Anime", result.Name)
	initialCallCount := suite.mockClient.GetByIDCallCount()

	result2, err := suite.animeService.GetAnimeByID(1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), result, result2)
	assert.Equal(suite.T(), initialCallCount, suite.mockClient.GetByIDCallCount())
}

func (suite *AnimeServiceIntegrationSuite) TestAddToFavoritesWithCache() {
	userID := int64(123)
	username := "testuser"

	suite.animeService.EnsureUserExists(userID, username)

	anime := &models.Anime{
		ID:      1001,
		Name:    "Favorite Anime",
		Russian: "Любимое Аниме",
		Image: models.AnimeImage{
			Preview: "/preview.jpg",
		},
	}

	err := suite.animeService.AddToFavorites(userID, *anime)
	assert.NoError(suite.T(), err)

	isFavorite, err := suite.animeService.IsFavorite(userID, 1001)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isFavorite)

	favorites, err := suite.animeService.GetUserFavorites(userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(favorites))
	assert.Equal(suite.T(), "Любимое Аниме", favorites[0].Title)
}

func (suite *AnimeServiceIntegrationSuite) TestRatingWithValidation() {
	userID := int64(123)
	username := "testuser"
	animeID := 1001

	suite.animeService.EnsureUserExists(userID, username)

	err := suite.animeService.AddRating(userID, animeID, 8)
	assert.NoError(suite.T(), err)

	rating, err := suite.animeService.GetUserRating(userID, animeID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), rating)
	assert.Equal(suite.T(), 8, rating.Score)

	err = suite.animeService.AddRating(userID, animeID, 0)
	assert.Error(suite.T(), err)

	err = suite.animeService.AddRating(userID, animeID, 11)
	assert.Error(suite.T(), err)
}

func (suite *AnimeServiceIntegrationSuite) TestDeleteRating() {
	userID := int64(123)
	username := "testuser"
	animeID := 1001

	suite.animeService.EnsureUserExists(userID, username)
	suite.animeService.AddRating(userID, animeID, 8)

	rating, err := suite.animeService.GetUserRating(userID, animeID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), rating)

	err = suite.animeService.DeleteRating(userID, animeID)
	assert.NoError(suite.T(), err)

	rating, err = suite.animeService.GetUserRating(userID, animeID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), rating)
}

func (suite *AnimeServiceIntegrationSuite) TestFullUserFlow() {
	userID := int64(456)
	username := "flowuser"

	err := suite.animeService.EnsureUserExists(userID, username)
	assert.NoError(suite.T(), err)

	mockAnimes := suite.getMockAnimes()
	query := "flow_test"
	suite.mockClient.SetSearchResults(query, mockAnimes)

	results, err := suite.animeService.SearchAnime(query)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), results)

	selectedAnime := results[0]

	err = suite.animeService.AddToFavorites(userID, selectedAnime)
	assert.NoError(suite.T(), err)

	err = suite.animeService.AddRating(userID, selectedAnime.ID, 9)
	assert.NoError(suite.T(), err)

	isFavorite, err := suite.animeService.IsFavorite(userID, selectedAnime.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isFavorite)

	rating, err := suite.animeService.GetUserRating(userID, selectedAnime.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 9, rating.Score)

	favorites, err := suite.animeService.GetUserFavorites(userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(favorites))
	assert.Equal(suite.T(), selectedAnime.ID, favorites[0].AnimeID)
}

func (suite *AnimeServiceIntegrationSuite) TestCountFavorites() {
	userID := int64(789)
	username := "countuser"

	suite.animeService.EnsureUserExists(userID, username)

	animeIDs := []int{1001, 1002, 1003}
	for i, id := range animeIDs {
		anime := &models.Anime{ID: id, Name: "Anime" + string(rune('1'+i))}
		suite.animeService.AddToFavorites(userID, *anime)
	}

	count, err := suite.animeService.CountFavorites(userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)
}

func (suite *AnimeServiceIntegrationSuite) getMockAnimes() []models.Anime {
	return []models.Anime{
		{
			ID:          1001,
			Name:        "Naruto",
			Russian:     "Наруто",
			Score:       "8.5",
			Status:      "released",
			Episodes:    220,
			Description: "Ninja adventures",
			Genres: []models.Genre{
				{ID: 1, Name: "action", Russian: "Экшен"},
			},
			Image: models.AnimeImage{
				Original: "https://example.com/1.jpg",
				Preview:  "/preview/1.jpg",
			},
		},
		{
			ID:          1002,
			Name:        "Demon Slayer",
			Russian:     "Истребитель демонов",
			Score:       "8.7",
			Status:      "released",
			Episodes:    26,
			Description: "Demon slaying",
			Genres: []models.Genre{
				{ID: 1, Name: "action", Russian: "Экшен"},
			},
			Image: models.AnimeImage{
				Original: "https://example.com/2.jpg",
				Preview:  "/preview/2.jpg",
			},
		},
	}
}

func TestAnimeServiceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AnimeServiceIntegrationSuite))
}
