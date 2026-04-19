package link

import (
	"encoding/json"
	"errors"
	"link-storage-service/internal/domain/link"
	"link-storage-service/internal/domain/response"
	"link-storage-service/internal/storage"
	"log/slog"
	"net/http"
	"time"
)

type StatsResponse struct {
	ShortCode string    `json:"short_code"`
	Url       string    `json:"url"`
	Visits    int64     `json:"visits"`
	CreatedAt time.Time `json:"created_at"`
}

type UrlStatsGetter interface {
	GetStats(shortCode string) (link.Stats, error)
}

func Stats(urlStatsGetter UrlStatsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")
		stats, err := urlStatsGetter.GetStats(shortCode)

		if errors.Is(err, storage.ErrUrlNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "link not found"})
			return
		}
		if err != nil {
			slog.Error("failed to get link with", "shortCode", shortCode)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "failed to to get link"})
			return
		}

		json.NewEncoder(w).Encode(StatsResponse{ShortCode: stats.ShortCode, Url: stats.Url, Visits: stats.Visits, CreatedAt: stats.CreatedAt})
	}
}
