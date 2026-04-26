package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"link-storage-service/internal/config"
	"link-storage-service/internal/domain/link"
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
	_, err = db.Exec(`
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
	return &Storage{db: db}, nil
}

func (s *Storage) SaveUrl(urlToSave, shortCode string) error {
	const op = "storage.postgres.SaveUrl"
	var id string
	err := s.db.QueryRow("INSERT INTO link(short_code, original_url) VALUES ($1, $2) RETURNING id", shortCode, urlToSave).Scan(&id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetAndIncrement(shortCode string) (link.SimpleLink, error) {
	const op = "storage.postgres.GetUrl"
	var resp link.SimpleLink
	err := s.db.QueryRow("UPDATE link SET visits = visits + 1 WHERE short_code = $1 RETURNING original_url, visits", shortCode).Scan(&resp.Url, &resp.Visits)
	if errors.Is(err, sql.ErrNoRows) {
		return link.SimpleLink{}, storage.ErrUrlNotFound
	}
	if err != nil {
		return link.SimpleLink{}, fmt.Errorf("%s: %w", op, err)
	}
	return resp, nil
}

func (s *Storage) IncrementVisits(shortCode string) (int64, error) {
	const op = "storage.postgres.IncrementVisits"
	var visits int64
	err := s.db.QueryRow("UPDATE link SET visits = visits + 1 WHERE short_code = $1 RETURNING visits", shortCode).Scan(&visits)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return visits, nil
}

func (s *Storage) DeleteUrl(shortCode string) error {
	const op = "storage.postgres.DeleteUrl"
	_, err := s.db.Exec("DELETE FROM link WHERE short_code = $1", shortCode)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetStats(shortCode string) (link.Stats, error) {
	const op = "storage.postgres.GetStats"
	var resp link.Stats
	err := s.db.QueryRow("SELECT short_code, original_url, visits, created_at FROM link WHERE short_code=$1", shortCode).Scan(&resp.ShortCode, &resp.Url, &resp.Visits, &resp.CreatedAt)
	if err != nil {
		return link.Stats{}, fmt.Errorf("%s: %w", op, err)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return link.Stats{}, storage.ErrUrlNotFound
	}
	return resp, nil
}

func (s *Storage) GetBatch(limit, offset int) ([]link.SimpleLink, error) {
	const op = "storage.postgres.GetBatch"
	rows, err := s.db.Query("SELECT original_url, visits FROM link LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return []link.SimpleLink{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	var resp []link.SimpleLink
	for rows.Next() {
		var u link.SimpleLink
		err := rows.Scan(&u.Url, &u.Visits)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		resp = append(resp, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if len(resp) == 0 {
		return []link.SimpleLink{}, nil
	}

	return resp, nil
}
