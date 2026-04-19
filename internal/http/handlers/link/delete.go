package link

import (
	"encoding/json"
	"link-storage-service/internal/cache"
	"link-storage-service/internal/domain/link"
	"link-storage-service/internal/domain/response"
	"log/slog"
	"net/http"
)

type UrlRemover interface {
	DeleteUrl(shortCode string) error
}

func Delete(urlRemover UrlRemover, cache *cache.Cache[link.SimpleLink]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")
		err := urlRemover.DeleteUrl(shortCode)
		if err != nil {
			slog.Error("failed to delete link with", "shortCode", shortCode)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "failed to to delete link"})
			return
		}
		cache.Delete(shortCode)
	}
}
