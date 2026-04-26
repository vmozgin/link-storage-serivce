package service

import (
	"context"
	"errors"
	"link-storage-service/internal/domain/link"
	"link-storage-service/internal/storage"
	"link-storage-service/internal/util/random"
	"log/slog"
	"strconv"
)

type Repository interface {
	SaveUrl(urlToSave, shortCode string) error
	IncrementVisits(shortCode string) (int64, error)
	GetAndIncrement(shortCode string) (link.SimpleLink, error)
	GetStats(shortCode string) (link.Stats, error)
	DeleteUrl(shortCode string) error
	GetBatch(limit, offset int) ([]link.SimpleLink, error)
}

type Cache interface {
	Set(ctx context.Context, shortCode string, url string) error
	Get(ctx context.Context, shortCode string) (string, error)
	Delete(ctx context.Context, shortCode string) error
}

type LinkService struct {
	repository Repository
	cache      Cache
}

func NewLinkService(repository Repository, cache Cache) *LinkService {
	return &LinkService{repository: repository, cache: cache}
}

func (s *LinkService) Create(url string) (string, error) {
	shortCode, err := random.Generate(6)
	if err != nil {
		slog.Error("failed to generate code", slog.String("error", err.Error()))
		return "", err
	}

	err = s.repository.SaveUrl(url, shortCode)
	if err != nil {
		slog.Error("failed to save url", slog.String("error", err.Error()))
		return "", err
	}

	return shortCode, nil
}

func (s *LinkService) Get(ctx context.Context, shortCode string) (link.SimpleLink, error) {
	cashedLink, err := s.cache.Get(ctx, shortCode)
	if err == nil {
		incrementedVisits, _ := s.repository.IncrementVisits(shortCode)
		return link.SimpleLink{Url: cashedLink, Visits: incrementedVisits}, nil
	}
	slog.Info("Link will be received form db")
	simpleLink, err := s.repository.GetAndIncrement(shortCode)
	if errors.Is(err, storage.ErrUrlNotFound) {
		return link.SimpleLink{}, err
	}
	if err != nil {
		return link.SimpleLink{}, errors.New("failed to get link")
	}
	s.cache.Set(ctx, shortCode, simpleLink.Url)

	return simpleLink, nil
}

func (s *LinkService) Stats(shortCode string) (link.Stats, error) {
	return s.repository.GetStats(shortCode)
}

func (s *LinkService) Delete(ctx context.Context, shortCode string) error {
	err := s.repository.DeleteUrl(shortCode)
	if err != nil {
		slog.Error("failed to delete link with", "shortCode", shortCode)
		return err
	}
	s.cache.Delete(ctx, shortCode)
	return nil
}

func (s *LinkService) GetBatch(limitStr, offsetStr string) ([]link.SimpleLink, error) {
	limit := 10
	offset := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}
	simpleLinks, err := s.repository.GetBatch(limit, offset)

	if err != nil {
		slog.Error("failed to get links")
		return []link.SimpleLink{}, err
	}

	return simpleLinks, nil
}
