package models

import "time"

type User struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	CreatedAt time.Time `db:"created_at"`
}

type Favorite struct {
	ID        int       `db:"id"`
	UserID    int64     `db:"user_id"`
	AnimeID   int       `db:"anime_id"`
	Title     string    `db:"title"`
	PosterURL string    `db:"poster_url"`
	AddedAt   time.Time `db:"added_at"`
}

type Rating struct {
	ID      int       `db:"id"`
	UserID  int64     `db:"user_id"`
	AnimeID int       `db:"anime_id"`
	Score   int       `db:"score"`
	RatedAt time.Time `db:"rated_at"`
}
