package link

import (
	"encoding/json"
	"link-storage-service/internal/domain/response"
	"link-storage-service/internal/service"
	"net/http"
)

func GetAll(linkService *service.LinkService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		simpleLinks, err := linkService.GetBatch(limitStr, offsetStr)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "failed to to get links"})
			return
		}
		json.NewEncoder(w).Encode(simpleLinks)
	}
}
