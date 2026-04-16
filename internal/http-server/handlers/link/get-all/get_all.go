package get_all

import (
	"encoding/json"
	"link-storage-service/internal/model/link"
	"link-storage-service/internal/model/response"
	"log/slog"
	"net/http"
	"strconv"
)

type UrlAllGetter interface {
	GetBatch(limit, offset int) ([]link.SimpleLink, error)
}

func New(urlAllGetter UrlAllGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		limit := 10
		offset := 1
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil {
				offset = o
			}
		}
		simpleLinks, err := urlAllGetter.GetBatch(limit, offset)

		if err != nil {
			slog.Error("failed to get links")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "failed to to get links"})
			return
		}
		json.NewEncoder(w).Encode(simpleLinks)
	}
}
