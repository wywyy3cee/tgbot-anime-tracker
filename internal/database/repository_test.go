package database

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
)

func newTestRepo(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	sx := sqlx.NewDb(db, "postgres")
	d := &Database{DB: sx}
	return NewRepository(d), mock
}

func TestAddFavorite_Success(t *testing.T) {
	repo, mock := newTestRepo(t)
	now := time.Now()
	fav := models.Favorite{
		UserID:    123,
		AnimeID:   1,
		Title:     "Death Note",
		PosterURL: "https://example.com/poster.jpg",
		AddedAt:   now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO favorites (user_id, anime_id, title, poster_url, added_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, anime_id) DO NOTHING
	`)).
		WithArgs(fav.UserID, fav.AnimeID, fav.Title, fav.PosterURL, now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AddFavorite(fav)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAddFavorite_DatabaseError(t *testing.T) {
	repo, mock := newTestRepo(t)
	fav := models.Favorite{
		UserID:  123,
		AnimeID: 1,
		Title:   "Test",
		AddedAt: time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO favorites (user_id, anime_id, title, poster_url, added_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, anime_id) DO NOTHING
	`)).
		WillReturnError(sql.ErrConnDone)

	err := repo.AddFavorite(fav)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestRemoveFavorite_Success(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	animeID := 1

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM favorites WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveFavorite(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRemoveFavorite_NotFound(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(999)
	animeID := 999

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM favorites WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.RemoveFavorite(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetFavorites_Success(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)

	rows := sqlmock.NewRows([]string{"id", "user_id", "anime_id", "title", "poster_url", "added_at"}).
		AddRow(1, userID, 1, "Death Note", "poster1.jpg", time.Now()).
		AddRow(2, userID, 2, "Naruto", "poster2.jpg", time.Now()).
		AddRow(3, userID, 3, "One Piece", "poster3.jpg", time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, anime_id, title, poster_url, added_at 
		FROM favorites 
		WHERE user_id = $1 
		ORDER BY added_at DESC
	`)).
		WithArgs(userID).
		WillReturnRows(rows)

	favorites, err := repo.GetFavorites(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(favorites) != 3 {
		t.Errorf("expected 3 favorites, got %d", len(favorites))
	}

	if favorites[0].Title != "Death Note" {
		t.Errorf("expected first favorite to be 'Death Note', got '%s'", favorites[0].Title)
	}
}

func TestGetFavorites_Empty(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(999)

	rows := sqlmock.NewRows([]string{"id", "user_id", "anime_id", "title", "poster_url", "added_at"})

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, anime_id, title, poster_url, added_at 
		FROM favorites 
		WHERE user_id = $1 
		ORDER BY added_at DESC
	`)).
		WithArgs(userID).
		WillReturnRows(rows)

	favorites, err := repo.GetFavorites(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(favorites) != 0 {
		t.Errorf("expected 0 favorites, got %d", len(favorites))
	}
}

func TestGetFavorites_DatabaseError(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, anime_id, title, poster_url, added_at 
		FROM favorites 
		WHERE user_id = $1 
		ORDER BY added_at DESC
	`)).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetFavorites(userID)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestIsFavorite_True(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	animeID := 1

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND anime_id = $2)`)).
		WithArgs(userID, animeID).
		WillReturnRows(rows)

	exists, err := repo.IsFavorite(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !exists {
		t.Error("expected anime to be in favorites")
	}
}

func TestIsFavorite_False(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	animeID := 999

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND anime_id = $2)`)).
		WithArgs(userID, animeID).
		WillReturnRows(rows)

	exists, err := repo.IsFavorite(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exists {
		t.Error("expected anime to not be in favorites")
	}
}

func TestCountFavorites_Success(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	expectedCount := 5

	rows := sqlmock.NewRows([]string{"count"}).AddRow(expectedCount)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM favorites WHERE user_id = $1`)).
		WithArgs(userID).
		WillReturnRows(rows)

	count, err := repo.CountFavorites(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != expectedCount {
		t.Errorf("expected count %d, got %d", expectedCount, count)
	}
}

func TestCountFavorites_Zero(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(999)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM favorites WHERE user_id = $1`)).
		WithArgs(userID).
		WillReturnRows(rows)

	count, err := repo.CountFavorites(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestCreateUser_Success(t *testing.T) {
	repo, mock := newTestRepo(t)
	now := time.Now()
	user := models.User{
		ID:        123,
		Username:  "test_user",
		CreatedAt: now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO users (id, username, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`)).
		WithArgs(user.ID, user.Username, now).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.CreateUser(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateUser_Duplicate(t *testing.T) {
	repo, mock := newTestRepo(t)
	user := models.User{
		ID:        123,
		Username:  "existing_user",
		CreatedAt: time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO users (id, username, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`)).
		WithArgs(user.ID, user.Username, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.CreateUser(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateUser_DatabaseError(t *testing.T) {
	repo, mock := newTestRepo(t)
	user := models.User{
		ID:        123,
		Username:  "test_user",
		CreatedAt: time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO users (id, username, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`)).
		WillReturnError(sql.ErrConnDone)

	err := repo.CreateUser(user)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestGetUser_Success(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "username", "created_at"}).
		AddRow(userID, "test_user", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, created_at FROM users WHERE id = $1`)).
		WithArgs(userID).
		WillReturnRows(rows)

	user, err := repo.GetUser(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user == nil {
		t.Error("expected non-nil user")
	}

	if user.ID != userID || user.Username != "test_user" {
		t.Errorf("unexpected user data: %v", user)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(999)

	rows := sqlmock.NewRows([]string{"id", "username", "created_at"})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, created_at FROM users WHERE id = $1`)).
		WithArgs(userID).
		WillReturnRows(rows)

	_, err := repo.GetUser(userID)
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestGetUser_DatabaseError(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, created_at FROM users WHERE id = $1`)).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetUser(userID)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestAddRating_Success(t *testing.T) {
	repo, mock := newTestRepo(t)
	now := time.Now()
	rating := models.Rating{
		UserID:  123,
		AnimeID: 1,
		Score:   8,
		RatedAt: now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO ratings (user_id, anime_id, score, rated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, anime_id) 
		DO UPDATE SET score = EXCLUDED.score, rated_at = EXCLUDED.rated_at
	`)).
		WithArgs(rating.UserID, rating.AnimeID, rating.Score, now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AddRating(rating)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAddRating_Update(t *testing.T) {
	repo, mock := newTestRepo(t)
	now := time.Now()
	rating := models.Rating{
		UserID:  123,
		AnimeID: 1,
		Score:   9,
		RatedAt: now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO ratings (user_id, anime_id, score, rated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, anime_id) 
		DO UPDATE SET score = EXCLUDED.score, rated_at = EXCLUDED.rated_at
	`)).
		WithArgs(rating.UserID, rating.AnimeID, rating.Score, now).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.AddRating(rating)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetRating_Success(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	animeID := 1
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "anime_id", "score", "rated_at"}).
		AddRow(1, userID, animeID, 8, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, anime_id, score, rated_at FROM ratings WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnRows(rows)

	rating, err := repo.GetRating(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rating == nil {
		t.Error("expected non-nil rating")
	}

	if rating.Score != 8 {
		t.Errorf("expected score 8, got %d", rating.Score)
	}
}

func TestGetRating_NotFound(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	animeID := 999

	rows := sqlmock.NewRows([]string{"id", "user_id", "anime_id", "score", "rated_at"})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, anime_id, score, rated_at FROM ratings WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnRows(rows)

	rating, err := repo.GetRating(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rating != nil {
		t.Error("expected nil for non-existent rating")
	}
}

func TestDeleteRating_Success(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	animeID := 1

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM ratings WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteRating(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteRating_NotFound(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(999)
	animeID := 999

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM ratings WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.DeleteRating(userID, animeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddRating_DatabaseError(t *testing.T) {
	repo, mock := newTestRepo(t)
	rating := models.Rating{
		UserID:  123,
		AnimeID: 1,
		Score:   8,
		RatedAt: time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO ratings (user_id, anime_id, score, rated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, anime_id) 
		DO UPDATE SET score = EXCLUDED.score, rated_at = EXCLUDED.rated_at
	`)).
		WillReturnError(sql.ErrConnDone)

	err := repo.AddRating(rating)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestDeleteRating_DatabaseError(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	animeID := 1

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM ratings WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnError(sql.ErrConnDone)

	err := repo.DeleteRating(userID, animeID)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestRemoveFavorite_DatabaseError(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	animeID := 1

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM favorites WHERE user_id = $1 AND anime_id = $2`)).
		WithArgs(userID, animeID).
		WillReturnError(sql.ErrConnDone)

	err := repo.RemoveFavorite(userID, animeID)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestIsFavorite_DatabaseError(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	animeID := 1

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND anime_id = $2)`)).
		WithArgs(userID, animeID).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.IsFavorite(userID, animeID)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestCountFavorites_DatabaseError(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM favorites WHERE user_id = $1`)).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.CountFavorites(userID)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestMultipleFavorites_Order(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := int64(123)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "anime_id", "title", "poster_url", "added_at"}).
		AddRow(3, userID, 3, "One Piece", "poster3.jpg", now).
		AddRow(2, userID, 2, "Naruto", "poster2.jpg", now.Add(-time.Hour)).
		AddRow(1, userID, 1, "Death Note", "poster1.jpg", now.Add(-2*time.Hour))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, anime_id, title, poster_url, added_at 
		FROM favorites 
		WHERE user_id = $1 
		ORDER BY added_at DESC
	`)).
		WithArgs(userID).
		WillReturnRows(rows)

	favorites, err := repo.GetFavorites(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(favorites) != 3 {
		t.Errorf("expected 3 favorites, got %d", len(favorites))
	}

	if favorites[0].Title != "One Piece" {
		t.Errorf("expected first to be 'One Piece', got '%s'", favorites[0].Title)
	}
}
