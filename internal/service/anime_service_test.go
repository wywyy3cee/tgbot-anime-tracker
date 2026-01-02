package service

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/database"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

type mockShikimoriClient struct {
	searchAnimeFunc func(query string, limit int) ([]models.Anime, error)
	getAnimeFunc    func(id int) (*models.Anime, error)
}

func (m *mockShikimoriClient) SearchAnime(query string, limit int) ([]models.Anime, error) {
	if m.searchAnimeFunc != nil {
		return m.searchAnimeFunc(query, limit)
	}
	return nil, errors.New("not implemented")
}

func (m *mockShikimoriClient) GetAnimeById(id int) (*models.Anime, error) {
	if m.getAnimeFunc != nil {
		return m.getAnimeFunc(id)
	}
	return nil, errors.New("not implemented")
}

type mockCache struct {
	getAnimeSearchFunc  func(query string) ([]models.Anime, error)
	setAnimeSearchFunc  func(query string, animes []models.Anime, duration time.Duration) error
	getAnimeDetailsFunc func(id int) (*models.Anime, error)
	setAnimeDetailsFunc func(id int, anime *models.Anime, duration time.Duration) error
}

func (m *mockCache) GetAnimeSearch(query string) ([]models.Anime, error) {
	if m.getAnimeSearchFunc != nil {
		return m.getAnimeSearchFunc(query)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCache) SetAnimeSearch(query string, animes []models.Anime, duration time.Duration) error {
	if m.setAnimeSearchFunc != nil {
		return m.setAnimeSearchFunc(query, animes, duration)
	}
	return nil
}

func (m *mockCache) GetAnimeDetails(id int) (*models.Anime, error) {
	if m.getAnimeDetailsFunc != nil {
		return m.getAnimeDetailsFunc(id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCache) SetAnimeDetails(id int, anime *models.Anime, duration time.Duration) error {
	if m.setAnimeDetailsFunc != nil {
		return m.setAnimeDetailsFunc(id, anime, duration)
	}
	return nil
}

func newTestService(t *testing.T) (*AnimeService, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	sx := sqlx.NewDb(db, "postgres")
	dbConn := &database.Database{DB: sx}
	repo := database.NewRepository(dbConn)

	service := &AnimeService{
		shikimoriClient: nil,
		repository:      repo,
		cache:           nil,
	}

	return service, mock
}

func newTestServiceWithMocks(t *testing.T) (*AnimeService, sqlmock.Sqlmock, *mockShikimoriClient, *mockCache) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	sx := sqlx.NewDb(db, "postgres")
	dbConn := &database.Database{DB: sx}
	repo := database.NewRepository(dbConn)

	shikimoriMock := &mockShikimoriClient{}
	cacheMock := &mockCache{}

	service := &AnimeService{
		shikimoriClient: shikimoriMock,
		repository:      repo,
		cache:           cacheMock,
	}

	return service, mock, shikimoriMock, cacheMock
}

func TestAddToFavorites_WithValidAnime(t *testing.T) {
	service, mock := newTestService(t)

	anime := models.Anime{
		ID:      1,
		Name:    "Death Note",
		Russian: "Тетрадь смерти",
		Image: models.AnimeImage{
			Preview: "/123.jpg",
		},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO favorites (user_id, anime_id, title, poster_url, added_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, anime_id) DO NOTHING
	`)).
		WithArgs(int64(123), anime.ID, "Тетрадь смерти", "https://shikimori.one/123.jpg", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := service.AddToFavorites(123, anime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchAnime_CacheHit(t *testing.T) {
	service, mock, _, cacheMock := newTestServiceWithMocks(t)

	expectedAnimes := []models.Anime{
		{ID: 1, Name: "Naruto", Description: "Anime about ninja"},
		{ID: 2, Name: "One Piece", Description: "Anime about pirates"},
	}

	cacheMock.getAnimeSearchFunc = func(query string) ([]models.Anime, error) {
		return expectedAnimes, nil
	}

	result, err := service.SearchAnime("naruto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != len(expectedAnimes) {
		t.Errorf("expected %d animes, got %d", len(expectedAnimes), len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchAnime_CacheMiss(t *testing.T) {
	service, mock, shikimoriMock, cacheMock := newTestServiceWithMocks(t)

	expectedAnimes := []models.Anime{
		{ID: 1, Name: "Naruto", Description: ""},
		{ID: 2, Name: "One Piece", Description: ""},
	}

	cacheMock.getAnimeSearchFunc = func(query string) ([]models.Anime, error) {
		return nil, errors.New("cache miss")
	}

	shikimoriMock.searchAnimeFunc = func(query string, limit int) ([]models.Anime, error) {
		return expectedAnimes, nil
	}

	cacheMock.setAnimeSearchFunc = func(query string, animes []models.Anime, duration time.Duration) error {
		return nil
	}

	result, err := service.SearchAnime("naruto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != len(expectedAnimes) {
		t.Errorf("expected %d animes, got %d", len(expectedAnimes), len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchAnime_ShikimoriError(t *testing.T) {
	service, mock, shikimoriMock, cacheMock := newTestServiceWithMocks(t)

	cacheMock.getAnimeSearchFunc = func(query string) ([]models.Anime, error) {
		return nil, errors.New("cache miss")
	}

	shikimoriMock.searchAnimeFunc = func(query string, limit int) ([]models.Anime, error) {
		return nil, errors.New("api error")
	}

	_, err := service.SearchAnime("naruto")
	if err == nil {
		t.Error("expected error from shikimori")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetAnimeByID_CacheHit(t *testing.T) {
	service, mock, _, cacheMock := newTestServiceWithMocks(t)

	expectedAnime := &models.Anime{ID: 1, Name: "Naruto", Description: "Anime about ninja"}

	cacheMock.getAnimeDetailsFunc = func(id int) (*models.Anime, error) {
		if id == 1 {
			return expectedAnime, nil
		}
		return nil, errors.New("not found")
	}

	result, err := service.GetAnimeByID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != expectedAnime.ID {
		t.Errorf("expected anime ID %d, got %d", expectedAnime.ID, result.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetAnimeByID_CacheMiss(t *testing.T) {
	service, mock, shikimoriMock, cacheMock := newTestServiceWithMocks(t)

	expectedAnime := &models.Anime{ID: 1, Name: "Naruto", Description: "Anime about ninja"}

	cacheMock.getAnimeDetailsFunc = func(id int) (*models.Anime, error) {
		return nil, errors.New("cache miss")
	}

	shikimoriMock.getAnimeFunc = func(id int) (*models.Anime, error) {
		if id == 1 {
			return expectedAnime, nil
		}
		return nil, errors.New("not found")
	}

	cacheMock.setAnimeDetailsFunc = func(id int, anime *models.Anime, duration time.Duration) error {
		return nil
	}

	result, err := service.GetAnimeByID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != expectedAnime.ID {
		t.Errorf("expected anime ID %d, got %d", expectedAnime.ID, result.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetAnimeByID_ShikimoriError(t *testing.T) {
	service, mock, shikimoriMock, cacheMock := newTestServiceWithMocks(t)

	cacheMock.getAnimeDetailsFunc = func(id int) (*models.Anime, error) {
		return nil, errors.New("cache miss")
	}

	shikimoriMock.getAnimeFunc = func(id int) (*models.Anime, error) {
		return nil, errors.New("api error")
	}

	_, err := service.GetAnimeByID(1)
	if err == nil {
		t.Error("expected error from shikimori")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchAnime_EnrichError(t *testing.T) {
	service, mock, shikimoriMock, cacheMock := newTestServiceWithMocks(t)

	cacheMock.getAnimeSearchFunc = func(query string) ([]models.Anime, error) {
		return nil, errors.New("cache miss")
	}

	animes := []models.Anime{
		{ID: 1, Name: "Naruto", Description: ""},
	}

	shikimoriMock.searchAnimeFunc = func(query string, limit int) ([]models.Anime, error) {
		return animes, nil
	}

	shikimoriMock.getAnimeFunc = func(id int) (*models.Anime, error) {
		return nil, errors.New("failed to get details")
	}

	result, err := service.SearchAnime("naruto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result[0].Description != "" {
		t.Errorf("expected empty description, got '%s'", result[0].Description)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRemoveFromFavorites(t *testing.T) {
	service, mock := newTestService(t)

	userID := int64(123)
	animeID := 1

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM favorites WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.RemoveFromFavorites(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetUserFavorites(t *testing.T) {
	service, mock := newTestService(t)

	userID := int64(123)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "anime_id", "title", "poster_url", "added_at"}).
		AddRow(1, userID, 1, "Death Note", "poster.jpg", now)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, anime_id, title, poster_url, added_at 
		FROM favorites 
		WHERE user_id = $1 
		ORDER BY added_at DESC
	`)).
		WithArgs(userID).
		WillReturnRows(rows)

	favorites, err := service.GetUserFavorites(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(favorites) != 1 {
		t.Errorf("expected 1 favorite, got %d", len(favorites))
	}

	if favorites[0].Title != "Death Note" {
		t.Errorf("expected title 'Death Note', got '%s'", favorites[0].Title)
	}
}

func TestIsFavorite_True(t *testing.T) {
	service, mock := newTestService(t)

	userID := int64(123)
	animeID := 1

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND anime_id = $2)`)).
		WithArgs(userID, animeID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := service.IsFavorite(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !exists {
		t.Error("expected anime to be in favorites")
	}
}

func TestCountFavorites(t *testing.T) {
	service, mock := newTestService(t)

	userID := int64(123)
	expectedCount := 5

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM favorites WHERE user_id = $1`)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

	count, err := service.CountFavorites(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != expectedCount {
		t.Errorf("expected count %d, got %d", expectedCount, count)
	}
}

func TestEnsureUserExists(t *testing.T) {
	service, mock := newTestService(t)

	userID := int64(789)
	username := "testuser"

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id, username, created_at) VALUES ($1, $2, $3)`)).
		WithArgs(userID, username, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := service.EnsureUserExists(userID, username)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAddRating_ValidScore(t *testing.T) {
	service, mock := newTestService(t)

	userID := int64(123)
	animeID := 5
	score := 8

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO ratings (user_id, anime_id, score, rated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, anime_id) DO UPDATE
		SET score = EXCLUDED.score, rated_at = EXCLUDED.rated_at
	`)).
		WithArgs(userID, animeID, score, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := service.AddRating(userID, animeID, score)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAddRating_InvalidScoreLow(t *testing.T) {
	service, _ := newTestService(t)

	err := service.AddRating(123, 1, 0)
	if err == nil {
		t.Error("expected error for score 0")
	}
}

func TestAddRating_InvalidScoreHigh(t *testing.T) {
	service, _ := newTestService(t)

	err := service.AddRating(123, 1, 11)
	if err == nil {
		t.Error("expected error for score 11")
	}
}

func TestGetUserRating(t *testing.T) {
	service, mock := newTestService(t)

	userID := int64(123)
	animeID := 5
	expectedScore := 8

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, anime_id, score, rated_at FROM ratings WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "anime_id", "score", "rated_at"}).
			AddRow(1, userID, animeID, expectedScore, time.Now()))

	rating, err := service.GetUserRating(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rating == nil || rating.Score != expectedScore {
		t.Errorf("expected rating with score %d, got %v", expectedScore, rating)
	}
}

func TestDeleteRating(t *testing.T) {
	service, mock := newTestService(t)

	userID := int64(123)
	animeID := 1

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM ratings WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.DeleteRating(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAddToFavorites_WithEnglishName(t *testing.T) {
	service, mock := newTestService(t)

	anime := models.Anime{
		ID:   2,
		Name: "Naruto",
		Image: models.AnimeImage{
			Preview: "/456.jpg",
		},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO favorites (user_id, anime_id, title, poster_url, added_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, anime_id) DO NOTHING
	`)).
		WithArgs(int64(456), anime.ID, "Naruto", "https://shikimori.one/456.jpg", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := service.AddToFavorites(456, anime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
