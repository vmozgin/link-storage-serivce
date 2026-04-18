package save

import (
	"encoding/json"
	"link-storage-service/internal/model/response"
	"link-storage-service/internal/util/random"
	"log/slog"
	"net/http"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ShortCode string `json:"short_code"`
}

type UrlSaver interface {
	SaveUrl(urlToSave, shortCode string) (string, error)
}

func New(urlSaver UrlSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			slog.Error("failed to decode request body", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "failed to decode request"})
			return
		}

		if req.URL == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "url is required"})
			return
		}

		shortCode, err := random.Generate(6)
		if err != nil {
			slog.Error("failed to generate code", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id, err := urlSaver.SaveUrl(req.URL, shortCode)
		if err != nil {
			slog.Error("failed to save url", slog.String("error", err.Error()))
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "failed to save url"})
			return
		}
		slog.Info("url saved", slog.String("id", id))
		json.NewEncoder(w).Encode(Response{ShortCode: shortCode})
	}
}
