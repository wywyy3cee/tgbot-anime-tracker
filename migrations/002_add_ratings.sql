-- +goose Up
CREATE TABLE IF NOT EXISTS ratings (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    anime_id INTEGER NOT NULL,
    score INTEGER NOT NULL CHECK (score >= 1 AND score <= 10),
    rated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, anime_id)
);

CREATE INDEX IF NOT EXISTS idx_ratings_user_id ON ratings(user_id);
CREATE INDEX IF NOT EXISTS idx_ratings_anime_id ON ratings(anime_id);

-- +goose Down
DROP INDEX IF EXISTS idx_ratings_anime_id;
DROP INDEX IF EXISTS idx_ratings_user_id;
DROP TABLE IF EXISTS ratings;