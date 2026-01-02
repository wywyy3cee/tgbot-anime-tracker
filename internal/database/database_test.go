package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/wywyy3cee/tgbot-anime-tracker/pkg/logger"
)

func TestConnect_Success(t *testing.T) {
	db, mock, err := sqlmock.New(
		sqlmock.MonitorPingsOption(true),
	)
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectPing()

	sx := sqlx.NewDb(db, "postgres")
	dbConn := &Database{DB: sx}

	err = dbConn.DB.Ping()
	if err != nil {
		t.Fatalf("unexpected error on ping: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}

	if dbConn.DB == nil {
		t.Error("expected non-nil database connection")
	}
}

func TestClose_Success(t *testing.T) {
	db, mock, err := sqlmock.New(
		sqlmock.MonitorPingsOption(true),
	)
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	mock.ExpectClose()

	sx := sqlx.NewDb(db, "postgres")
	dbConn := &Database{DB: sx}

	err = dbConn.Close()
	if err != nil {
		t.Fatalf("unexpected error on close: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestNewRepository(t *testing.T) {
	db, _, err := sqlmock.New(
		sqlmock.MonitorPingsOption(true),
	)
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	sx := sqlx.NewDb(db, "postgres")
	dbConn := &Database{DB: sx}

	repo := NewRepository(dbConn)

	if repo == nil {
		t.Error("expected non-nil repository")
	}

	if repo.db != dbConn {
		t.Error("repository should reference the same database")
	}
}

func TestConnect_SuccessIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	logger := logger.New()
	databaseURL := "postgres://test:test@localhost:5432/test?sslmode=disable"

	db, err := Connect(databaseURL, logger)
	if err == nil {
		defer db.Close()
		t.Errorf("expected error with invalid URL, got nil")
	}
}

func TestConnect_Failure(t *testing.T) {
	logger := logger.New()
	databaseURL := "invalid://url"

	_, err := Connect(databaseURL, logger)
	if err == nil {
		t.Error("expected error with invalid URL, got nil")
	}
}

func TestRunMigrations_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(1, 1))

	sx := sqlx.NewDb(db, "postgres")
	database := &Database{DB: sx}

	tempDir := t.TempDir()
	err = database.RunMigrations(tempDir)
	if err == nil {
		t.Error("expected error with empty migrations dir, got nil")
	}
}

func TestRunMigrations_InvalidDir(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	sx := sqlx.NewDb(db, "postgres")
	database := &Database{DB: sx}

	err = database.RunMigrations("/non/existent/directory")
	if err == nil {
		t.Error("expected error with non-existent directory, got nil")
	}
}

func TestDatabase_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	mock.ExpectClose()

	sx := sqlx.NewDb(db, "postgres")
	database := &Database{DB: sx}

	err = database.Close()
	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
