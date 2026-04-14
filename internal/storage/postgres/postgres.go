package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"link-storage-service/internal/config"
	"link-storage-service/internal/storage"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(cfg config.Storage) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS link(
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    	short_code VARCHAR(32) NOT NULL UNIQUE,
    	original_url TEXT NOT NULL,
    	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    	visits BIGINT NOT NULL DEFAULT 0)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUrl(urlToSave, shortCode string) (string, error) {
	const op = "storage.postgres.SaveUrl"

	stmt, err := s.db.Prepare("INSERT INTO link(short_code, original_url) VALUES ($1, $2) RETURNING id")

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var id string
	err = stmt.QueryRow(shortCode, urlToSave).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUrl(shortCode string) (string, error) {
	const op = "storage.postgres.GetUrl"

	stmt, err := s.db.Prepare("SELECT original_url FROM link WHERE short_code = $1")

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var originalUrl string
	err = stmt.QueryRow(shortCode).Scan(&originalUrl)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrUrlNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return originalUrl, nil
}

func (s *Storage) DeleteUrl(shortCode string) error {
	const op = "storage.postgres.DeleteUrl"

	stmt, err := s.db.Prepare("DELETE FROM link WHERE short_code = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(shortCode)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
