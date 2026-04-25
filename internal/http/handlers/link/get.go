package link

import (
	"encoding/json"
	"errors"
	"link-storage-service/internal/cache"
	"link-storage-service/internal/domain/link"
	"link-storage-service/internal/domain/response"
	"link-storage-service/internal/storage"
	"log/slog"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type GetResponse struct {
	Url    string `json:"url"`
	Visits int64  `json:"visits"`
}

type UrlGetter interface {
	GetAndIncrement(shortCode string) (link.SimpleLink, error)
	IncrementVisits(shortCode string) (int64, error)
}

func Get(urlGetter UrlGetter, cache *cache.RedisCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")
		ctx := r.Context()
		cashedLink, err := cache.Get(ctx, shortCode)
		if !errors.Is(err, redis.Nil) {
			invrementedVisits, _ := urlGetter.IncrementVisits(shortCode)
			json.NewEncoder(w).Encode(GetResponse{Url: cashedLink, Visits: invrementedVisits})
			return
		}
		slog.Info("Link will be received form db")
		simpleLink, err := urlGetter.GetAndIncrement(shortCode)
		if errors.Is(err, storage.ErrUrlNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "link not found"})
			return
		}
		if err != nil {
			slog.Error("failed to get link with", "shortCode", shortCode)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "failed to get link"})
			return
		}
		cache.Set(ctx, shortCode, simpleLink.Url)
		json.NewEncoder(w).Encode(GetResponse{Url: simpleLink.Url, Visits: simpleLink.Visits})
	}
}
