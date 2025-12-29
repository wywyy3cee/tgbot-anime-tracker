package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wywyy3cee/tgbot-anime-tracker/internal/models"
	"github.com/wywyy3cee/tgbot-anime-tracker/pkg/logger"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
	logger *logger.Logger
}

func New(redisURL string, logger *logger.Logger) (*Cache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("Redis cache connected")

	return &Cache{
		client: client,
		ctx:    ctx,
		logger: logger,
	}, nil
}

func (c *Cache) GetAnimeSearch(query string) ([]models.Anime, error) {
	key := fmt.Sprintf("anime:search:%s", query)
	c.logger.Debug("Getting anime search from cache: %s", key)
	val, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		c.logger.Debug("Cache miss for search: %s", query)
		return nil, nil
	}
	if err != nil {
		c.logger.Error("Failed to get from cache: %v", err)
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var animes []models.Anime
	if err := json.Unmarshal([]byte(val), &animes); err != nil {
		c.logger.Error("Failed to unmarshal cached data: %v", err)
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	c.logger.Info("Cache hit for search: %s, found %d animes", query, len(animes))
	return animes, nil
}

func (c *Cache) SetAnimeSearch(query string, animes []models.Anime, ttl time.Duration) error {
	key := fmt.Sprintf("anime:search:%s", query)
	c.logger.Debug("Setting anime search in cache: %s, ttl: %v", key, ttl)
	data, err := json.Marshal(animes)
	if err != nil {
		c.logger.Error("Failed to marshal anime list: %v", err)
		return fmt.Errorf("failed to marshal anime list: %w", err)
	}

	if err := c.client.Set(c.ctx, key, data, ttl).Err(); err != nil {
		c.logger.Error("Failed to set cache: %v", err)
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.logger.Info("Cached search result for: %s, %d animes", query, len(animes))
	return nil
}

func (c *Cache) GetAnimeDetails(id int) (*models.Anime, error) {
	key := fmt.Sprintf("anime:details:%d", id)
	c.logger.Debug("Getting anime details from cache: %s", key)
	val, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		c.logger.Debug("Cache miss for anime ID: %d", id)
		return nil, nil
	}
	if err != nil {
		c.logger.Error("Failed to get from cache: %v", err)
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var anime models.Anime
	if err := json.Unmarshal([]byte(val), &anime); err != nil {
		c.logger.Error("Failed to unmarshal cached data: %v", err)
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	c.logger.Info("Cache hit for anime ID: %d", id)
	return &anime, nil
}

func (c *Cache) SetAnimeDetails(id int, anime *models.Anime, ttl time.Duration) error {
	key := fmt.Sprintf("anime:details:%d", id)
	c.logger.Debug("Setting anime details in cache: %s, ttl: %v", key, ttl)
	data, err := json.Marshal(anime)
	if err != nil {
		c.logger.Error("Failed to marshal anime: %v", err)
		return fmt.Errorf("failed to marshal anime: %w", err)
	}

	if err := c.client.Set(c.ctx, key, data, ttl).Err(); err != nil {
		c.logger.Error("Failed to set cache: %v", err)
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.logger.Info("Cached anime details for ID: %d", id)
	return nil
}

func (c *Cache) Close() error {
	return c.client.Close()
}
