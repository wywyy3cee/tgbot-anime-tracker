-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    username VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS favorites (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    anime_id INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    poster_url TEXT,
    added_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, anime_id)
);

CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_anime_id ON favorites(anime_id);

-- +goose Down
DROP INDEX IF EXISTS idx_favorites_anime_id;
DROP INDEX IF EXISTS idx_favorites_user_id;
DROP TABLE IF EXISTS favorites;
DROP TABLE IF EXISTS users;