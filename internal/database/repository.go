package database

import (
	"fmt"

	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

type Repository struct {
	db *Database
}

func NewRepository(db *Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(user models.User) error {
	query := `
		INSERT INTO users (id, username, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := r.db.DB.Exec(query, user.ID, user.Username, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *Repository) GetUser(userID int64) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, created_at FROM users WHERE id = $1`

	err := r.db.DB.Get(&user, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *Repository) AddFavorite(favorite models.Favorite) error {
	query := `
		INSERT INTO favorites (user_id, anime_id, title, poster_url, added_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, anime_id) DO NOTHING
	`
	_, err := r.db.DB.Exec(query,
		favorite.UserID,
		favorite.AnimeID,
		favorite.Title,
		favorite.PosterURL,
		favorite.AddedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add favorite: %w", err)
	}
	return nil
}

func (r *Repository) RemoveFavorite(userID int64, animeID int) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND anime_id = $2`

	_, err := r.db.DB.Exec(query, userID, animeID)
	if err != nil {
		return fmt.Errorf("failed to remove favorite: %w", err)
	}
	return nil
}

func (r *Repository) GetFavorites(userID int64) ([]models.Favorite, error) {
	var favorites []models.Favorite
	query := `
		SELECT id, user_id, anime_id, title, poster_url, added_at 
		FROM favorites 
		WHERE user_id = $1 
		ORDER BY added_at DESC
	`

	err := r.db.DB.Select(&favorites, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorites: %w", err)
	}
	return favorites, nil
}

func (r *Repository) IsFavorite(userID int64, animeID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND anime_id = $2)`

	err := r.db.DB.Get(&exists, query, userID, animeID)
	if err != nil {
		return false, fmt.Errorf("failed to check favorite: %w", err)
	}
	return exists, nil
}

func (r *Repository) CountFavorites(userID int64) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`

	err := r.db.DB.Get(&count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count favorites: %w", err)
	}
	return count, nil
}

func (r *Repository) AddRating(rating models.Rating) error {
	query := `
		INSERT INTO ratings (user_id, anime_id, score, rated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, anime_id) 
		DO UPDATE SET score = EXCLUDED.score, rated_at = EXCLUDED.rated_at
	`
	_, err := r.db.DB.Exec(query,
		rating.UserID,
		rating.AnimeID,
		rating.Score,
		rating.RatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add rating: %w", err)
	}
	return nil
}

func (r *Repository) GetRating(userID int64, animeID int) (*models.Rating, error) {
	var rating models.Rating
	query := `SELECT id, user_id, anime_id, score, rated_at FROM ratings WHERE user_id = $1 AND anime_id = $2`

	err := r.db.DB.Get(&rating, query, userID, animeID)
	if err != nil {
		return nil, nil
	}
	return &rating, nil
}

func (r *Repository) DeleteRating(userID int64, animeID int) error {
	query := `DELETE FROM ratings WHERE user_id = $1 AND anime_id = $2`

	_, err := r.db.DB.Exec(query, userID, animeID)
	if err != nil {
		return fmt.Errorf("failed to delete rating: %w", err)
	}
	return nil
}
