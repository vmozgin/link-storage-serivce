package delete

import (
	"encoding/json"
	"link-storage-service/internal/model/response"
	"log/slog"
	"net/http"
)

type UrlRemover interface {
	DeleteUrl(shortCode string) error
}

func New(urlRemover UrlRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")
		err := urlRemover.DeleteUrl(shortCode)
		if err != nil {
			slog.Error("failed to delete link with", "shortCode", shortCode)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "failed to to delete link"})
			return
		}
	}
}
