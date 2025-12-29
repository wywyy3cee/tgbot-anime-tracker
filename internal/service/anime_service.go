package service

import (
	"fmt"
	"time"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/cache"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/database"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/shikimori"
)

type AnimeService struct {
	shikimoriClient *shikimori.Client
	repository      *database.Repository
	cache           *cache.Cache
}

func NewAnimeService(client *shikimori.Client, repo *database.Repository, cache *cache.Cache) *AnimeService {
	return &AnimeService{
		shikimoriClient: client,
		repository:      repo,
		cache:           cache,
	}
}

func (s *AnimeService) SearchAnime(query string) ([]models.Anime, error) {
	if s.cache != nil {
		cached, err := s.cache.GetAnimeSearch(query)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	animes, err := s.shikimoriClient.SearchAnime(query, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search anime: %w", err)
	}

	if s.cache != nil {
		_ = s.cache.SetAnimeSearch(query, animes, time.Hour)
	}

	return animes, nil
}

func (s *AnimeService) GetAnimeByID(id int) (*models.Anime, error) {
	if s.cache != nil {
		cached, err := s.cache.GetAnimeDetails(id)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	anime, err := s.shikimoriClient.GetAnimeById(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime by id: %w", err)
	}

	if s.cache != nil {
		_ = s.cache.SetAnimeDetails(id, anime, 24*time.Hour)
	}

	return anime, nil
}

func (s *AnimeService) AddToFavorites(userID int64, anime models.Anime) error {
	title := anime.Russian
	if title == "" {
		title = anime.Name
	}

	posterURL := ""
	if anime.Image.Preview != "" {
		posterURL = "https://shikimori.one" + anime.Image.Preview
	}

	favorite := models.Favorite{
		UserID:    userID,
		AnimeID:   anime.ID,
		Title:     title,
		PosterURL: posterURL,
		AddedAt:   time.Now(),
	}

	return s.repository.AddFavorite(favorite)
}

func (s *AnimeService) RemoveFromFavorites(userID int64, animeID int) error {
	return s.repository.RemoveFavorite(userID, animeID)
}

func (s *AnimeService) GetUserFavorites(userID int64) ([]models.Favorite, error) {
	return s.repository.GetFavorites(userID)
}

func (s *AnimeService) IsFavorite(userID int64, animeID int) (bool, error) {
	return s.repository.IsFavorite(userID, animeID)
}

func (s *AnimeService) CountFavorites(userID int64) (int, error) {
	return s.repository.CountFavorites(userID)
}

func (s *AnimeService) EnsureUserExists(userID int64, username string) error {
	user := models.User{
		ID:        userID,
		Username:  username,
		CreatedAt: time.Now(),
	}
	return s.repository.CreateUser(user)
}
