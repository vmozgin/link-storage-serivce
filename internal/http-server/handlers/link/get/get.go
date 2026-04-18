package get

import (
	"encoding/json"
	"errors"
	"link-storage-service/internal/cache"
	"link-storage-service/internal/model/link"
	"link-storage-service/internal/model/response"
	"link-storage-service/internal/storage"
	"log/slog"
	"net/http"
)

type Response struct {
	Url    string `json:"url"`
	Visits int64  `json:"visits"`
}

type UrlGetter interface {
	GetAndIncrement(shortCode string) (link.SimpleLink, error)
}

func New(urlGetter UrlGetter, cash *cache.Cache[link.SimpleLink]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")
		cashedLink, ok := cash.Get(shortCode)
		if ok {
			json.NewEncoder(w).Encode(Response{Url: cashedLink.Url, Visits: cashedLink.Visits})
			return
		}
		slog.Info("Link will be received form db")
		simpleUrl, err := urlGetter.GetAndIncrement(shortCode)
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
		cash.Set(shortCode, simpleUrl)
		json.NewEncoder(w).Encode(Response{Url: simpleUrl.Url, Visits: simpleUrl.Visits})
	}
}
