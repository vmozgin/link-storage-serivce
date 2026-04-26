package link

import (
	"encoding/json"
	"errors"
	"link-storage-service/internal/domain/response"
	"link-storage-service/internal/service"
	"link-storage-service/internal/storage"
	"log/slog"
	"net/http"
)

type GetResponse struct {
	Url    string `json:"url"`
	Visits int64  `json:"visits"`
}

func Get(linkService *service.LinkService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")
		ctx := r.Context()
		link, err := linkService.Get(ctx, shortCode)
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
		json.NewEncoder(w).Encode(GetResponse{Url: link.Url, Visits: link.Visits})
	}
}
