package link

import (
	"encoding/json"
	"link-storage-service/internal/domain/response"
	"link-storage-service/internal/service"
	"net/http"
)

func Delete(linkService *service.LinkService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")
		err := linkService.Delete(r.Context(), shortCode)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Error: "failed to to delete link"})
			return
		}
	}
}
