package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
	"github.com/wywyy3cee/tgbot-anime-tracker/pkg/logger"
)

func TestGetAnimeSearch_CacheHit(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	expected := []models.Anime{
		{ID: 1, Name: "Death Note", Russian: "Тетрадь смерти", Score: "9.0"},
		{ID: 2, Name: "Naruto", Russian: "Наруто", Score: "8.5"},
	}
	data, _ := json.Marshal(expected)
	mock.ExpectGet("anime:search:death").SetVal(string(data))

	result, err := c.GetAnimeSearch("death")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 || result[0].ID != 1 {
		t.Errorf("expected correct anime data, got %v", result)
	}
}

func TestGetAnimeSearch_CacheMiss(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	mock.ExpectGet("anime:search:nonexistent").SetErr(redis.Nil)

	result, err := c.GetAnimeSearch("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil for cache miss, got %v", result)
	}
}

func TestGetAnimeSearch_InvalidJSON(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	mock.ExpectGet("anime:search:bad").SetVal("not valid json")

	_, err := c.GetAnimeSearch("bad")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGetAnimeSearch_RedisError(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	mock.ExpectGet("anime:search:error").SetErr(redis.Nil)

	result, err := c.GetAnimeSearch("error")
	if err != nil {
		t.Fatalf("unexpected error for Nil: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil on Nil error, got %v", result)
	}
}

func TestSetAnimeSearch_Success(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	animes := []models.Anime{
		{ID: 1, Name: "One Piece", Russian: "Ван Пис", Score: "8.9"},
	}
	data, _ := json.Marshal(animes)
	mock.ExpectSet("anime:search:onepie", data, time.Hour).SetVal("OK")

	err := c.SetAnimeSearch("onepie", animes, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetAnimeSearch_DifferentTTL(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	animes := []models.Anime{{ID: 1, Name: "Test"}}
	data, _ := json.Marshal(animes)

	mock.ExpectSet("anime:search:test", data, 24*time.Hour).SetVal("OK")

	err := c.SetAnimeSearch("test", animes, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetAnimeSearch_EmptyList(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	animes := []models.Anime{}
	data, _ := json.Marshal(animes)
	mock.ExpectSet("anime:search:empty", data, time.Hour).SetVal("OK")

	err := c.SetAnimeSearch("empty", animes, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetAnimeSearch_RedisError(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	animes := []models.Anime{{ID: 1, Name: "Test"}}
	data, _ := json.Marshal(animes)
	mock.ExpectSet("anime:search:error", data, time.Hour).SetErr(redis.Nil)

	err := c.SetAnimeSearch("error", animes, time.Hour)
	if err == nil {
		t.Error("expected error from Redis")
	}
}

func TestGetAnimeDetails_CacheHit(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	expected := &models.Anime{
		ID:          1,
		Name:        "Death Note",
		Russian:     "Тетрадь смерти",
		Score:       "9.0",
		Status:      "finished",
		Episodes:    37,
		Description: "Psychological thriller",
	}
	data, _ := json.Marshal(expected)
	mock.ExpectGet("anime:details:1").SetVal(string(data))

	result, err := c.GetAnimeDetails(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil || result.ID != 1 || result.Name != "Death Note" {
		t.Errorf("expected correct anime details, got %v", result)
	}
}

func TestGetAnimeDetails_CacheMiss(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	mock.ExpectGet("anime:details:999").SetErr(redis.Nil)

	result, err := c.GetAnimeDetails(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil for cache miss, got %v", result)
	}
}

func TestSetAnimeDetails_Success(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	anime := &models.Anime{
		ID:          2,
		Name:        "Naruto",
		Russian:     "Наруто",
		Score:       "8.5",
		Description: "Long-running ninja anime",
	}
	data, _ := json.Marshal(anime)
	mock.ExpectSet("anime:details:2", data, 24*time.Hour).SetVal("OK")

	err := c.SetAnimeDetails(2, anime, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetAnimeDetails_LongTTL(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	anime := &models.Anime{ID: 3, Name: "One Piece"}
	data, _ := json.Marshal(anime)
	mock.ExpectSet("anime:details:3", data, 24*time.Hour).SetVal("OK")

	err := c.SetAnimeDetails(3, anime, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetAnimeDetails_InvalidJSON(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	mock.ExpectGet("anime:details:bad").SetVal("corrupted json }{")

	_, err := c.GetAnimeDetails(1)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestSetAnimeDetails_InvalidJSON(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	type UnmarshalableAnime struct {
		Ch chan int // Каналы не могут быть маршализированы в JSON
	}

	// Это заставит json.Marshal вернуть ошибку

	anime := &models.Anime{ID: 1, Name: "Test"}
	data, _ := json.Marshal(anime)
	mock.ExpectSet("anime:details:1", data, time.Hour).SetVal("OK")

	err := c.SetAnimeDetails(1, anime, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCacheKeyFormat(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	animes := []models.Anime{{ID: 1, Name: "Test"}}
	data, _ := json.Marshal(animes)

	// anime:search:{query}
	mock.ExpectSet("anime:search:test_query", data, time.Hour).SetVal("OK")
	err := c.SetAnimeSearch("test_query", animes, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// anime:details:{id}
	anime := &models.Anime{ID: 42, Name: "Test"}
	data, _ = json.Marshal(anime)
	mock.ExpectSet("anime:details:42", data, time.Hour).SetVal("OK")
	err = c.SetAnimeDetails(42, anime, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMultipleCacheOperations(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	searchAnimes := []models.Anime{{ID: 1, Name: "Test1"}}
	searchData, _ := json.Marshal(searchAnimes)
	mock.ExpectSet("anime:search:query1", searchData, time.Hour).SetVal("OK")

	detailAnime := &models.Anime{ID: 1, Name: "DetailTest1"}
	detailData, _ := json.Marshal(detailAnime)
	mock.ExpectSet("anime:details:1", detailData, time.Hour).SetVal("OK")

	mock.ExpectGet("anime:search:query1").SetVal(string(searchData))

	err := c.SetAnimeSearch("query1", searchAnimes, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.SetAnimeDetails(1, detailAnime, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := c.GetAnimeSearch("query1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 anime, got %d", len(result))
	}
}

func TestGetAnimeDetails_RedisError(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	mock.ExpectGet("anime:details:500").SetErr(redis.Nil)

	result, err := c.GetAnimeDetails(500)
	if err != nil {
		t.Fatalf("unexpected error for Nil: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil on error, got %v", result)
	}
}

func TestSetAnimeDetails_RedisError(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	anime := &models.Anime{ID: 10, Name: "Test"}
	data, _ := json.Marshal(anime)
	mock.ExpectSet("anime:details:10", data, time.Hour).SetErr(redis.Nil)

	err := c.SetAnimeDetails(10, anime, time.Hour)
	if err == nil {
		t.Error("expected error from Redis")
	}
}

func TestClose(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	err := c.Close()
	if err != nil {
		t.Fatalf("unexpected error on close: %v", err)
	}
}

func TestGetAnimeSearch_LargeQueryString(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	largeQuery := "very long anime name with many characters that could be used in search query"
	expectedAnimes := []models.Anime{
		{ID: 1, Name: "Anime 1", Score: "9.0"},
	}
	data, _ := json.Marshal(expectedAnimes)

	key := "anime:search:" + largeQuery
	mock.ExpectGet(key).SetVal(string(data))

	result, err := c.GetAnimeSearch(largeQuery)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 anime, got %d", len(result))
	}
}

func TestSetAnimeSearch_LargeDataset(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	var largeAnimeList []models.Anime
	for i := 0; i < 100; i++ {
		largeAnimeList = append(largeAnimeList, models.Anime{
			ID:    i + 1,
			Name:  "Anime " + string(rune(i)),
			Score: "8.0",
		})
	}

	data, _ := json.Marshal(largeAnimeList)
	mock.ExpectSet("anime:search:large", data, time.Hour).SetVal("OK")

	err := c.SetAnimeSearch("large", largeAnimeList, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetAnimeDetails_LargeDescription(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	largeDescription := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris. "

	expected := &models.Anime{
		ID:          1,
		Name:        "Test Anime",
		Description: largeDescription,
		Score:       "8.5",
	}
	data, _ := json.Marshal(expected)
	mock.ExpectGet("anime:details:1").SetVal(string(data))

	result, err := c.GetAnimeDetails(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Description != largeDescription {
		t.Errorf("expected full description, got shorter one")
	}
}

func TestSetAnimeDetails_WithAllFields(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	anime := &models.Anime{
		ID:          99,
		Name:        "Full Anime",
		Russian:     "Полное аниме",
		Kind:        "TV",
		Score:       "9.5",
		Status:      "finished",
		Episodes:    50,
		AiredOn:     "2020-01-01",
		ReleasedOn:  "2020-12-31",
		Description: "A complete anime with all fields",
		Image: models.AnimeImage{
			Original: "https://example.com/original.jpg",
			Preview:  "https://example.com/preview.jpg",
		},
		Genres: []models.Genre{
			{ID: 1, Name: "Action", Russian: "Экшен"},
			{ID: 2, Name: "Adventure", Russian: "Приключения"},
		},
	}

	data, _ := json.Marshal(anime)
	mock.ExpectSet("anime:details:99", data, 24*time.Hour).SetVal("OK")

	err := c.SetAnimeDetails(99, anime, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetAnimeSearch_SpecialCharacters(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	specialQuery := "anime:with@special#chars"
	expectedAnimes := []models.Anime{
		{ID: 1, Name: "Special Anime"},
	}
	data, _ := json.Marshal(expectedAnimes)

	key := "anime:search:" + specialQuery
	mock.ExpectGet(key).SetVal(string(data))

	result, err := c.GetAnimeSearch(specialQuery)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestSetAnimeSearch_WithZeroTTL(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	animes := []models.Anime{{ID: 1, Name: "Temporary"}}
	data, _ := json.Marshal(animes)
	mock.ExpectSet("anime:search:temp", data, 0).SetVal("OK")

	err := c.SetAnimeSearch("temp", animes, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnimeWithoutGenres(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	anime := &models.Anime{
		ID:     5,
		Name:   "No Genres Anime",
		Genres: []models.Genre{},
	}

	data, _ := json.Marshal(anime)
	mock.ExpectSet("anime:details:5", data, time.Hour).SetVal("OK")

	err := c.SetAnimeDetails(5, anime, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCacheConsistency(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	originalAnime := &models.Anime{
		ID:    7,
		Name:  "Test Consistency",
		Score: "8.0",
	}

	data, _ := json.Marshal(originalAnime)
	mock.ExpectSet("anime:details:7", data, time.Hour).SetVal("OK")
	mock.ExpectGet("anime:details:7").SetVal(string(data))

	// Set
	err := c.SetAnimeDetails(7, originalAnime, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error on set: %v", err)
	}

	// Get
	retrievedAnime, err := c.GetAnimeDetails(7)
	if err != nil {
		t.Fatalf("unexpected error on get: %v", err)
	}

	if retrievedAnime.ID != originalAnime.ID ||
		retrievedAnime.Name != originalAnime.Name ||
		retrievedAnime.Score != originalAnime.Score {
		t.Error("retrieved anime data does not match original")
	}
}

func TestCacheDifferentTTLValues(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	anime := &models.Anime{ID: 1, Name: "Test"}

	ttlValues := []time.Duration{
		time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
		time.Minute,
		10 * time.Minute,
	}

	for i, ttl := range ttlValues {
		animeID := 100 + i
		key := fmt.Sprintf("anime:details:%d", animeID)
		anime.ID = animeID

		data, _ := json.Marshal(anime)
		mock.ExpectSet(key, data, ttl).SetVal("OK")

		err := c.SetAnimeDetails(animeID, anime, ttl)
		if err != nil {
			t.Fatalf("unexpected error with TTL %v: %v", ttl, err)
		}
	}
}

func TestGetMultipleAnimeIds(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	logger := logger.New()
	c := &Cache{client: rdb, ctx: context.Background(), logger: logger}

	for i := 1; i <= 5; i++ {
		anime := &models.Anime{
			ID:   i,
			Name: "Anime " + string(rune(i)),
		}
		data, _ := json.Marshal(anime)
		key := fmt.Sprintf("anime:details:%d", i)
		mock.ExpectGet(key).SetVal(string(data))
	}

	for i := 1; i <= 5; i++ {
		result, err := c.GetAnimeDetails(i)
		if err != nil {
			t.Fatalf("unexpected error for ID %d: %v", i, err)
		}

		if result == nil || result.ID != i {
			t.Errorf("expected anime ID %d, got %v", i, result)
		}
	}
}
