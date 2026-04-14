package save

import (
	"encoding/json"
	resp "link-storage-service/internal/model/response"
	"link-storage-service/internal/util/random"
	"log/slog"
	"net/http"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	resp.Response
	ShortCode string `json:"shortCode"`
}

type UrlSaver interface {
	SaveUrl(urlToSave, shortCode string) (string, error)
}

func New(urlSaver UrlSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			slog.Error("failed to decode request body", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp.Error("failed to decode request"))
			return
		}

		if req.URL == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp.Error("url is required"))
			return
		}

		shortCode, err := random.Generate(6)
		if err != nil {
			slog.Error("failed to generate code", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		slog.Info("request body decoded", slog.Any("request", req))

		id, err := urlSaver.SaveUrl(req.URL, shortCode)
		if err != nil {
			slog.Error("failed to save url", slog.String("error", err.Error()))
			json.NewEncoder(w).Encode(resp.Error("failed to save url"))
			return
		}
		slog.Info("url saved", slog.String("id", id))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Response: resp.OK(), ShortCode: shortCode})
	}
}
