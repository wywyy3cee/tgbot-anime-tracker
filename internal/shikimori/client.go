package shikimori

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/ratelimit"
)

type Client struct {
	baseURL     string
	httpClient  *http.Client
	rateLimiter *ratelimit.RateLimiter
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		rateLimiter: ratelimit.NewRateLimiter(80), // 80 requests/minute (buffer 10)
	}
}

func (c *Client) SearchAnime(query string, limit int) ([]models.Anime, error) {
	ctx := context.Background()

	// wait for available token
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait error: %w", err)
	}

	encodedQuery := url.QueryEscape(query)
	endpoint := fmt.Sprintf("%s/animes?search=%s&limit=%d", c.baseURL, encodedQuery, limit)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "TelegramAnimeBot/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error get search of anime list: %w", err)
	}
	defer resp.Body.Close()

	// 429 Too Many Requests
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limit exceeded, try again later")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var animes []models.Anime
	err = json.NewDecoder(resp.Body).Decode(&animes)
	if err != nil {
		return nil, fmt.Errorf("error decoding anime list response: %w", err)
	}

	if len(animes) == 0 {
		return nil, fmt.Errorf("no animes found for query: %s", query)
	}

	return animes, nil
}

func (c *Client) GetAnimeById(id int) (*models.Anime, error) {
	ctx := context.Background()

	// wait for available token
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait error: %w", err)
	}

	endpoint := fmt.Sprintf("%s/animes/%d", c.baseURL, id)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "TelegramAnimeBot/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting anime by id: %w", err)
	}
	defer resp.Body.Close()

	// 429 Too Many Requests
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limit exceeded, try again later")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("anime with id %d not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var anime models.Anime
	err = json.NewDecoder(resp.Body).Decode(&anime)
	if err != nil {
		return nil, fmt.Errorf("error decoding anime response: %w", err)
	}

	return &anime, nil
}
