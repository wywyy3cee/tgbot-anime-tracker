package tests

import (
	"os"
	"path/filepath"
)

const (
	DefaultDBHost     = "db"
	DefaultDBPort     = "5432"
	DefaultDBUser     = "postgres"
	DefaultDBPassword = "postgres"
	DefaultDBName     = "anime_bot"
	DefaultDBSSLMode  = "disable"

	DefaultRedisHost = "redis"
	DefaultRedisPort = "6379"
	DefaultRedisDB   = 0

	MigrationsPath = "./migrations"
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
	migrationsDir := getEnv("TEST_MIGRATIONS_DIR", "")
	if migrationsDir == "" {
		if _, err := os.Stat("./migrations"); err == nil {
			migrationsDir = "./migrations"
		} else if wd, err := os.Getwd(); err == nil {
			migrationsDir = filepath.Join(wd, "migrations")
			if _, err := os.Stat(migrationsDir); err != nil {
				migrationsDir = "./migrations"
			}
		}
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
	return "postgres://" + tc.DBUser + ":" + tc.DBPassword + "@" + tc.DBHost + ":" + tc.DBPort + "/" + tc.DBName + "?sslmode=" + tc.DBSSLMODE
}

func (tc *TestConfig) GetRedisURL() string {
	return "redis://" + tc.RedisHost + ":" + tc.RedisPort + "/" + string(rune('0'+tc.RedisDB))
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
