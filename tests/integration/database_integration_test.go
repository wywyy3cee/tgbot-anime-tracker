package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/database"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
	"github.com/wywyy3cee/tgbot-anime-tracker/pkg/logger"
	"github.com/wywyy3cee/tgbot-anime-tracker/tests"
)

type DatabaseIntegrationSuite struct {
	suite.Suite
	db         *database.Database
	repository *database.Repository
	logger     *logger.Logger
	testConfig *tests.TestConfig
}

func (suite *DatabaseIntegrationSuite) SetupSuite() {
	var err error

	suite.logger = logger.New()
	suite.testConfig = tests.NewTestConfig()

	dbURL := suite.testConfig.GetDatabaseURL()
	suite.db, err = database.Connect(dbURL, suite.logger)
	require.NoError(suite.T(), err, "Failed to connect to database")

	err = suite.db.RunMigrations(suite.testConfig.MigrationsDir)
	require.NoError(suite.T(), err, "Failed to run migrations")

	suite.repository = database.NewRepository(suite.db)
}

func (suite *DatabaseIntegrationSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *DatabaseIntegrationSuite) TearDownTest() {
	suite.db.DB.Exec("TRUNCATE TABLE users CASCADE")
	suite.db.DB.Exec("TRUNCATE TABLE favorites CASCADE")
	suite.db.DB.Exec("TRUNCATE TABLE ratings CASCADE")
}

func (suite *DatabaseIntegrationSuite) TestUserCreation() {
	user := models.User{
		ID:        12345,
		Username:  "testuser",
		CreatedAt: getNow(),
	}

	err := suite.repository.CreateUser(user)
	assert.NoError(suite.T(), err, "Failed to create user")

	retrieved, err := suite.repository.GetUser(user.ID)
	assert.NoError(suite.T(), err, "Failed to retrieve user")
	assert.NotNil(suite.T(), retrieved)
	assert.Equal(suite.T(), user.ID, retrieved.ID)
	assert.Equal(suite.T(), user.Username, retrieved.Username)
}

func (suite *DatabaseIntegrationSuite) TestUserCreationDuplicate() {
	user := models.User{
		ID:        12345,
		Username:  "testuser",
		CreatedAt: getNow(),
	}

	err := suite.repository.CreateUser(user)
	assert.NoError(suite.T(), err)

	err = suite.repository.CreateUser(user)
	assert.NoError(suite.T(), err, "Duplicate user creation should not fail")
}

func (suite *DatabaseIntegrationSuite) TestFavoriteOperations() {
	userID := int64(12345)
	username := "testuser"

	user := models.User{ID: userID, Username: username, CreatedAt: getNow()}
	suite.repository.CreateUser(user)

	favorite := models.Favorite{
		UserID:    userID,
		AnimeID:   1001,
		Title:     "Test Anime",
		PosterURL: "https://example.com/poster.jpg",
		AddedAt:   getNow(),
	}

	err := suite.repository.AddFavorite(favorite)
	assert.NoError(suite.T(), err, "Failed to add favorite")

	isFavorite, err := suite.repository.IsFavorite(userID, 1001)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isFavorite, "Anime should be in favorites")

	favorites, err := suite.repository.GetFavorites(userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(favorites), "Should have 1 favorite")
	assert.Equal(suite.T(), 1001, favorites[0].AnimeID)
	assert.Equal(suite.T(), "Test Anime", favorites[0].Title)

	count, err := suite.repository.CountFavorites(userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)

	err = suite.repository.RemoveFavorite(userID, 1001)
	assert.NoError(suite.T(), err, "Failed to remove favorite")

	isFavorite, err = suite.repository.IsFavorite(userID, 1001)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isFavorite, "Anime should not be in favorites")

	favorites, err = suite.repository.GetFavorites(userID)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), favorites, "Favorites list should be empty")
}

func (suite *DatabaseIntegrationSuite) TestMultipleFavorites() {
	userID := int64(12345)
	username := "testuser"

	user := models.User{ID: userID, Username: username, CreatedAt: getNow()}
	suite.repository.CreateUser(user)

	animeIDs := []int{1001, 1002, 1003, 1004}
	for i, animeID := range animeIDs {
		favorite := models.Favorite{
			UserID:    userID,
			AnimeID:   animeID,
			Title:     "Anime " + string(rune('A'+i)),
			PosterURL: "https://example.com/poster" + string(rune('A'+i)) + ".jpg",
			AddedAt:   getNow(),
		}
		err := suite.repository.AddFavorite(favorite)
		assert.NoError(suite.T(), err)
	}

	count, err := suite.repository.CountFavorites(userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 4, count)

	favorites, err := suite.repository.GetFavorites(userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 4, len(favorites))

	suite.repository.RemoveFavorite(userID, 1001)
	suite.repository.RemoveFavorite(userID, 1003)

	count, err = suite.repository.CountFavorites(userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count)
}

func (suite *DatabaseIntegrationSuite) TestRatingOperations() {
	userID := int64(12345)
	username := "testuser"
	animeID := 1001

	user := models.User{ID: userID, Username: username, CreatedAt: getNow()}
	suite.repository.CreateUser(user)

	rating := models.Rating{
		UserID:  userID,
		AnimeID: animeID,
		Score:   8,
		RatedAt: getNow(),
	}

	err := suite.repository.AddRating(rating)
	assert.NoError(suite.T(), err, "Failed to add rating")

	retrieved, err := suite.repository.GetRating(userID, animeID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrieved)
	assert.Equal(suite.T(), 8, retrieved.Score)

	updatedRating := models.Rating{
		UserID:  userID,
		AnimeID: animeID,
		Score:   9,
		RatedAt: getNow(),
	}

	err = suite.repository.AddRating(updatedRating)
	assert.NoError(suite.T(), err)

	retrieved, err = suite.repository.GetRating(userID, animeID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 9, retrieved.Score)

	err = suite.repository.DeleteRating(userID, animeID)
	assert.NoError(suite.T(), err)

	retrieved, err = suite.repository.GetRating(userID, animeID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), retrieved)
}

func (suite *DatabaseIntegrationSuite) TestFavoritesAndRatings() {
	userID := int64(12345)
	username := "testuser"

	user := models.User{ID: userID, Username: username, CreatedAt: getNow()}
	suite.repository.CreateUser(user)

	favorite := models.Favorite{
		UserID:    userID,
		AnimeID:   1001,
		Title:     "Test Anime",
		PosterURL: "https://example.com/poster.jpg",
		AddedAt:   getNow(),
	}
	suite.repository.AddFavorite(favorite)

	rating := models.Rating{
		UserID:  userID,
		AnimeID: 1001,
		Score:   8,
		RatedAt: getNow(),
	}
	suite.repository.AddRating(rating)

	isFavorite, _ := suite.repository.IsFavorite(userID, 1001)
	assert.True(suite.T(), isFavorite)

	userRating, _ := suite.repository.GetRating(userID, 1001)
	assert.NotNil(suite.T(), userRating)

	suite.repository.RemoveFavorite(userID, 1001)
	suite.repository.DeleteRating(userID, 1001)

	isFavorite, _ = suite.repository.IsFavorite(userID, 1001)
	assert.False(suite.T(), isFavorite)

	userRating, _ = suite.repository.GetRating(userID, 1001)
	assert.Nil(suite.T(), userRating)
}

func getNow() time.Time {
	return time.Now()
}

func TestDatabaseIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DatabaseIntegrationSuite))
}
