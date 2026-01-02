package shikimori

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

func TestSearchAnime_Success(t *testing.T) {
	expectedAnimes := []models.Anime{
		{ID: 1, Name: "Death Note", Russian: "Тетрадь смерти", Score: "9.0"},
		{ID: 2, Name: "Death Note Another Note", Russian: "ДН: Another Note", Score: "7.5"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "/animes") {
			t.Errorf("expected /animes path, got %s", r.URL.Path)
		}

		if !strings.Contains(r.URL.RawQuery, "search=Death") {
			t.Errorf("expected search parameter, got %s", r.URL.RawQuery)
		}

		if !strings.Contains(r.URL.RawQuery, "limit=10") {
			t.Errorf("expected limit parameter, got %s", r.URL.RawQuery)
		}

		if r.Header.Get("User-Agent") != "TelegramAnimeBot/1.0" {
			t.Errorf("expected correct User-Agent, got %s", r.Header.Get("User-Agent"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedAnimes)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.SearchAnime("Death Note", 10)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 animes, got %d", len(result))
	}

	if result[0].ID != 1 || result[0].Name != "Death Note" {
		t.Errorf("expected Death Note anime, got %v", result[0])
	}
}

func TestSearchAnime_NoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.Anime{})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.SearchAnime("NonExistentAnime", 10)

	if err == nil {
		t.Error("expected error for no results")
	}

	if !strings.Contains(err.Error(), "no animes found") {
		t.Errorf("expected 'no animes found' error, got: %v", err)
	}
}

func TestSearchAnime_RateLimitExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.SearchAnime("Test", 10)

	if err == nil {
		t.Error("expected error for rate limit")
	}

	if !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}

func TestSearchAnime_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.SearchAnime("Test", 10)

	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "error decoding") {
		t.Errorf("expected decoding error, got: %v", err)
	}
}

func TestSearchAnime_UnexpectedStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.SearchAnime("Test", 10)

	if err == nil {
		t.Error("expected error for unexpected status code")
	}

	if !strings.Contains(err.Error(), "unexpected status code") {
		t.Errorf("expected status code error, got: %v", err)
	}
}

func TestSearchAnime_SpecialCharactersInQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "search=") {
			t.Logf("Query: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.Anime{{ID: 1, Name: "Test"}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.SearchAnime("日本", 10)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected search results")
	}
}

func TestGetAnimeById_Success(t *testing.T) {
	expectedAnime := &models.Anime{
		ID:          1,
		Name:        "Death Note",
		Russian:     "Тетрадь смерти",
		Score:       "9.0",
		Status:      "finished",
		Episodes:    37,
		Description: "Psychological thriller about a notebook that kills",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/animes/1") {
			t.Errorf("expected /animes/1 path, got %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedAnime)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetAnimeById(1)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil || result.ID != 1 {
		t.Errorf("expected anime with ID 1, got %v", result)
	}

	if result.Name != "Death Note" {
		t.Errorf("expected name 'Death Note', got '%s'", result.Name)
	}
}

func TestGetAnimeById_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetAnimeById(99999)

	if err == nil {
		t.Error("expected error for non-existent anime")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestGetAnimeById_RateLimitExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetAnimeById(1)

	if err == nil {
		t.Error("expected error for rate limit")
	}

	if !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}

func TestGetAnimeById_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not a json"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetAnimeById(1)

	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "error decoding") {
		t.Errorf("expected decoding error, got: %v", err)
	}
}

func TestGetAnimeById_UnexpectedStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetAnimeById(1)

	if err == nil {
		t.Error("expected error for unexpected status code")
	}

	if !strings.Contains(err.Error(), "unexpected status code") {
		t.Errorf("expected status code error, got: %v", err)
	}
}

func TestSearchAnime_LargeResponse(t *testing.T) {
	animes := make([]models.Anime, 100)
	for i := 0; i < 100; i++ {
		animes[i] = models.Anime{
			ID:   i + 1,
			Name: fmt.Sprintf("Anime %d", i+1),
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(animes)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.SearchAnime("Popular", 100)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 100 {
		t.Errorf("expected 100 animes, got %d", len(result))
	}
}

func TestGetAnimeById_PartialData(t *testing.T) {
	partialAnime := &models.Anime{
		ID:   5,
		Name: "Incomplete Anime",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(partialAnime)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetAnimeById(5)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil || result.ID != 5 {
		t.Errorf("expected anime with ID 5, got %v", result)
	}

	if result.Russian != "" {
		t.Errorf("expected empty Russian name, got '%s'", result.Russian)
	}
}

func TestNewClient_Configuration(t *testing.T) {
	client := NewClient("https://api.shikimori.one")

	if client == nil {
		t.Error("expected non-nil client")
	}

	if client.httpClient == nil {
		t.Error("expected initialized HTTP client")
	}

	if client.rateLimiter == nil {
		t.Error("expected initialized rate limiter")
	}

	expectedTimeout := 10
	if client.httpClient.Timeout.Seconds() != float64(expectedTimeout) {
		t.Errorf("expected timeout %d seconds, got %v", expectedTimeout, client.httpClient.Timeout)
	}
}
