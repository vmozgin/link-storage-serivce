package service

import (
	"context"
	"errors"
	"link-storage-service/internal/domain/link"
	"link-storage-service/internal/storage"
	"testing"
)

type mockRepository struct {
	Repository
	urlToSave             string
	shortCode             string
	saveUrlErr            error
	incrementedVisits     int64
	incrVisitsErr         error
	getAndIncrementCalled bool
	incrementVisitsCalled bool
	simpleLinkUrl         string
	simpleLinkVisits      int64
	getAndIncrementErr    error
}

func (m *mockRepository) SaveUrl(urlToSave, shortCode string) error {
	m.urlToSave = urlToSave
	m.shortCode = shortCode
	return m.saveUrlErr
}

func (m *mockRepository) IncrementVisits(shortCode string) (int64, error) {
	m.incrementVisitsCalled = true
	return m.incrementedVisits, m.incrVisitsErr
}

func (m *mockRepository) GetAndIncrement(shortCode string) (link.SimpleLink, error) {
	m.getAndIncrementCalled = true
	return link.SimpleLink{Url: m.simpleLinkUrl, Visits: m.incrementedVisits}, m.getAndIncrementErr
}

type mockCache struct {
	Cache
	getResult string
	getError  error
	setError  error
	setCalled bool
	shortCode string
	url       string
}

func (m *mockCache) Get(ctx context.Context, shortCode string) (string, error) {
	return m.getResult, m.getError
}

func (m *mockCache) Set(ctx context.Context, shortCode string, url string) error {
	m.setCalled = true
	m.shortCode = shortCode
	m.url = url
	return m.setError
}

const testURL = "https://example.com"
const testShortCode = "123456"

func TestCreate_Success(t *testing.T) {
	repo := &mockRepository{}
	cache := &mockCache{}
	svc := NewLinkService(repo, cache)

	shortCode, err := svc.Create(testURL)

	if err != nil {
		t.Fatalf("expected not error, but got %v", err)
	}
	if shortCode == "" || len(shortCode) != 6 {
		t.Fatalf("expected length 6, but got %v", len(shortCode))
	}
	if repo.urlToSave != testURL {
		t.Fatalf("saved testURL %v not equal to expectedUrl %v", repo.urlToSave, testURL)
	}
	if repo.shortCode != shortCode {
		t.Fatalf("saved shortCode %v not equal to expectedShortCode %v", repo.shortCode, shortCode)
	}
}

func TestCreate_RepositoryReturnError(t *testing.T) {
	repo := &mockRepository{saveUrlErr: errors.New("save error")}
	cache := &mockCache{}
	svc := NewLinkService(repo, cache)

	shortCode, err := svc.Create(testURL)

	if err == nil {
		t.Fatal("expected error, but has no error")
	}
	if shortCode != "" {
		t.Fatalf("expected empty string, but got %v", shortCode)
	}
}

func TestGetFromCache_Success(t *testing.T) {
	const testVisits = 5
	repo := &mockRepository{incrementedVisits: testVisits}
	cache := &mockCache{getResult: testURL}
	svc := NewLinkService(repo, cache)

	resultUrl, err := svc.Get(context.Background(), testShortCode)
	if err != nil {
		t.Fatalf("expected not error, but got %v", err)
	}
	if resultUrl.Url != testURL {
		t.Fatalf("expected resultUrl %v, but got %v", testURL, resultUrl.Url)
	}
	if resultUrl.Visits != testVisits {
		t.Fatalf("expected visits %v, but visits %v", testVisits, resultUrl.Visits)
	}
	if repo.getAndIncrementCalled {
		t.Fatal("getAndIncrement should not be called")
	}
	if cache.setCalled {
		t.Fatal("cache.Set should not be called")
	}
}

func TestGetFromDb_Success(t *testing.T) {
	const testVisits = 5
	repo := &mockRepository{simpleLinkUrl: testURL, incrementedVisits: testVisits}
	cache := &mockCache{getError: errors.New("not found cache")}
	svc := NewLinkService(repo, cache)

	resultUrl, err := svc.Get(context.Background(), testShortCode)
	if err != nil {
		t.Fatalf("expected not error, but got %v", err)
	}
	if resultUrl.Url != testURL {
		t.Fatalf("expected resultUrl %v, but got %v", testURL, resultUrl.Url)
	}
	if resultUrl.Visits != testVisits {
		t.Fatalf("expected visits %v, but visits %v", testVisits, resultUrl.Visits)
	}
	if cache.shortCode != testShortCode {
		t.Fatalf("expected shortCode %v, but got %v", testShortCode, cache.shortCode)
	}
	if cache.url != testURL {
		t.Fatalf("expected url %v, but got %v", testURL, cache.url)
	}
	if repo.incrementVisitsCalled {
		t.Fatal("IncrementVisits should not be called")
	}
}

func TestGetLinkNotFondFromDb(t *testing.T) {
	repo := &mockRepository{getAndIncrementErr: storage.ErrUrlNotFound}
	cache := &mockCache{getError: errors.New("not found cache")}
	svc := NewLinkService(repo, cache)

	resultUrl, err := svc.Get(context.Background(), testShortCode)
	if err == nil {
		t.Fatal("expected error, but not")
	}
	if !errors.Is(err, storage.ErrUrlNotFound) {
		t.Fatalf("expected ErrUrlNotFound, but got %v", err)
	}
	if resultUrl.Url != "" {
		t.Fatalf("expected url \"\", but got %v", resultUrl.Url)
	}
	if resultUrl.Visits != 0 {
		t.Fatalf("expected visits 0, but got %v", resultUrl.Visits)
	}
	if repo.incrementVisitsCalled {
		t.Fatal("IncrementVisits should not be called")
	}
	if cache.setCalled {
		t.Fatal("cache.Set should not be called")
	}
}
