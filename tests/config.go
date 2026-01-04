package tests

import (
	"os"
	"strconv"
)

const (
	DefaultDBHost     = "localhost"
	DefaultDBPort     = "5432"
	DefaultDBUser     = "postgres"
	DefaultDBPassword = "postgres"
	DefaultDBName     = "anime_bot"
	DefaultDBSSLMode  = "disable"
	DefaultRedisHost  = "redis"
	DefaultRedisPort  = "6379"
	DefaultRedisDB    = 0
)

type TestConfig struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMODE     string
	RedisHost     string
	RedisPort     string
	RedisDB       int
	MigrationsDir string
}

func NewTestConfig() *TestConfig {
	migrationsDir := os.Getenv("TEST_MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}

	return &TestConfig{
		DBHost:        getEnv("TEST_DB_HOST", DefaultDBHost),
		DBPort:        getEnv("TEST_DB_PORT", DefaultDBPort),
		DBUser:        getEnv("TEST_DB_USER", DefaultDBUser),
		DBPassword:    getEnv("TEST_DB_PASSWORD", DefaultDBPassword),
		DBName:        getEnv("TEST_DB_NAME", DefaultDBName),
		DBSSLMODE:     getEnv("TEST_DB_SSLMODE", DefaultDBSSLMode),
		RedisHost:     getEnv("TEST_REDIS_HOST", DefaultRedisHost),
		RedisPort:     getEnv("TEST_REDIS_PORT", DefaultRedisPort),
		RedisDB:       DefaultRedisDB,
		MigrationsDir: migrationsDir,
	}
}

func (tc *TestConfig) GetDatabaseURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		return dbURL
	}
	return "postgres://" + tc.DBUser + ":" + tc.DBPassword + "@" + tc.DBHost + ":" + tc.DBPort + "/" + tc.DBName + "?sslmode=" + tc.DBSSLMODE
}

func (tc *TestConfig) GetRedisURL() string {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		return redisURL
	}
	return "redis://" + tc.RedisHost + ":" + tc.RedisPort + "/" + strconv.Itoa(tc.RedisDB)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
