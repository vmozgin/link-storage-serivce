package get

import (
	"encoding/json"
	"errors"
	"link-storage-service/internal/model/response"
	"link-storage-service/internal/model/url"
	"link-storage-service/internal/storage"
	"log/slog"
	"net/http"
)

type Response struct {
	Url    string `json:"url"`
	Visits int64  `json:"visits"`
}

type UrlGetter interface {
	GetAndIncrement(shortCode string) (url.SimpleUrl, error)
}

func New(urlGetter UrlGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")
		simpleUrl, err := urlGetter.GetAndIncrement(shortCode)
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

		json.NewEncoder(w).Encode(Response{Url: simpleUrl.Url, Visits: simpleUrl.Visits})
	}
}
