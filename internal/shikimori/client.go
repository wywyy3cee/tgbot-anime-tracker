package shikimori

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c Client) SearchAnime(query string, limit int) ([]models.Anime, error) {
	encodedQuery := url.QueryEscape(query)
	endpoint := fmt.Sprintf("%s/animes?search=%s&limit=%d", c.baseURL, encodedQuery, limit)

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error get search of anime list: %w", err)
	}

	defer resp.Body.Close()
	resp.Header.Set("User-Agent", "TelegramAnimeBot/1.0")

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

func (c Client) GetAnimeById(id int) (*models.Anime, error) {
	endpoint := fmt.Sprintf("%s/animes/%d", c.baseURL, id)

	req, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error get anime by id: %w", err)
	}

	req.Header.Set("User-Agent", "TelegramAnimeBot/1.0")
	defer req.Body.Close()

	if req.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", req.StatusCode)
	}

	var anime models.Anime
	err = json.NewDecoder(req.Body).Decode(&anime)
	if err != nil {
		return nil, fmt.Errorf("error decoding anime response: %w", err)
	}

	return &anime, nil
}
