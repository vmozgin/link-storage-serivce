package link

import (
	"encoding/json"
	"link-storage-service/internal/domain/response"
	"link-storage-service/internal/service"
	"log/slog"
	"net/http"
)

type Request struct {
	URL string `json:"url"`
}

type CreateResponse struct {
	ShortCode string `json:"short_code"`
}

func Create(linkService *service.LinkService) http.HandlerFunc {
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

		shortCode, err := linkService.Create(req.URL)
		if err != nil {
			slog.Error("failed to create link", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "failed to create link"})
			return
		}
		json.NewEncoder(w).Encode(CreateResponse{ShortCode: shortCode})
	}
}
