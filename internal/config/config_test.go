package config

import (
	"os"
	"testing"
)

func TestLoad_AllEnvVarsSet(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost/dbname")
	os.Setenv("BOT_TOKEN", "1234567890:ABCDEfghijklmnopqrstuvwxyz")
	os.Setenv("REDIS_URL", "redis://localhost:6379/0")
	os.Setenv("SHIKIMORI_URL", "https://api.shikimori.one")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("SHIKIMORI_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if cfg.DatabaseURL != "postgres://user:pass@localhost/dbname" {
		t.Errorf("expected correct DatabaseURL, got '%s'", cfg.DatabaseURL)
	}

	if cfg.BotToken != "1234567890:ABCDEfghijklmnopqrstuvwxyz" {
		t.Errorf("expected correct BotToken, got '%s'", cfg.BotToken)
	}

	if cfg.RedisURL != "redis://localhost:6379/0" {
		t.Errorf("expected correct RedisURL, got '%s'", cfg.RedisURL)
	}

	if cfg.ShikimoriURL != "https://api.shikimori.one" {
		t.Errorf("expected correct ShikimoriURL, got '%s'", cfg.ShikimoriURL)
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Setenv("BOT_TOKEN", "token")
	os.Setenv("REDIS_URL", "redis://url")
	os.Setenv("SHIKIMORI_URL", "https://api")

	defer func() {
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("SHIKIMORI_URL")
	}()

	_, err := Load()
	if err == nil {
		t.Error("expected error when DATABASE_URL is missing")
	}
}

func TestLoad_MissingBotToken(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://url")
	os.Unsetenv("BOT_TOKEN")
	os.Setenv("REDIS_URL", "redis://url")
	os.Setenv("SHIKIMORI_URL", "https://api")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("SHIKIMORI_URL")
	}()

	_, err := Load()
	if err == nil {
		t.Error("expected error when BOT_TOKEN is missing")
	}
}

func TestLoad_MissingRedisURL(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://url")
	os.Setenv("BOT_TOKEN", "token")
	os.Unsetenv("REDIS_URL")
	os.Setenv("SHIKIMORI_URL", "https://api")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("SHIKIMORI_URL")
	}()

	_, err := Load()
	if err == nil {
		t.Error("expected error when REDIS_URL is missing")
	}
}

func TestLoad_MissingShikimoriURL(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://url")
	os.Setenv("BOT_TOKEN", "token")
	os.Setenv("REDIS_URL", "redis://url")
	os.Unsetenv("SHIKIMORI_URL")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("REDIS_URL")
	}()

	_, err := Load()
	if err == nil {
		t.Error("expected error when SHIKIMORI_URL is missing")
	}
}

func TestLoad_AllEnvVarsMissing(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("BOT_TOKEN")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("SHIKIMORI_URL")

	_, err := Load()
	if err == nil {
		t.Error("expected error when all environment variables are missing")
	}
}

func TestLoad_EmptyStringValues(t *testing.T) {
	os.Setenv("DATABASE_URL", "")
	os.Setenv("BOT_TOKEN", "token")
	os.Setenv("REDIS_URL", "redis://url")
	os.Setenv("SHIKIMORI_URL", "https://api")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("SHIKIMORI_URL")
	}()

	_, err := Load()
	if err == nil {
		t.Error("expected error when DATABASE_URL is empty string")
	}
}

func TestLoad_SpecialCharactersInValues(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/database?sslmode=require")
	os.Setenv("BOT_TOKEN", "123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZ-_")
	os.Setenv("REDIS_URL", "redis://:password@localhost:6379/0?db=0")
	os.Setenv("SHIKIMORI_URL", "https://api-test.shikimori.one/v2/animes")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("SHIKIMORI_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/database?sslmode=require" {
		t.Errorf("expected DatabaseURL to preserve special characters")
	}

	if cfg.BotToken != "123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZ-_" {
		t.Errorf("expected BotToken to preserve special characters")
	}
}

func TestLoad_LongValues(t *testing.T) {
	longURL := "postgres://user:password@host.subdomain.example.com:5432/very_long_database_name_with_underscores_and_numbers_12345?sslmode=require&timeout=30"
	longToken := "123456789012345678901234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	os.Setenv("DATABASE_URL", longURL)
	os.Setenv("BOT_TOKEN", longToken)
	os.Setenv("REDIS_URL", "redis://localhost:6379")
	os.Setenv("SHIKIMORI_URL", "https://api.shikimori.one")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("SHIKIMORI_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DatabaseURL != longURL {
		t.Errorf("expected long DatabaseURL to be preserved")
	}

	if cfg.BotToken != longToken {
		t.Errorf("expected long BotToken to be preserved")
	}
}

func TestLoad_MultipleConsecutiveCalls(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://url1")
	os.Setenv("BOT_TOKEN", "token1")
	os.Setenv("REDIS_URL", "redis://url1")
	os.Setenv("SHIKIMORI_URL", "https://api1")

	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("SHIKIMORI_URL")
	}()

	cfg1, err1 := Load()
	if err1 != nil {
		t.Fatalf("first load failed: %v", err1)
	}

	cfg2, err2 := Load()
	if err2 != nil {
		t.Fatalf("second load failed: %v", err2)
	}

	if cfg1.DatabaseURL != cfg2.DatabaseURL {
		t.Error("expected same config on consecutive loads")
	}
}
