package handler

import (
	"encoding/json"
	"generator/internal/service"
	"net/http"
	"strconv"
)

type ClipCandidateHandler struct {
	clipService *service.ClipCandidateService
}

func NewClipCandidateHandler(clipService *service.ClipCandidateService) *ClipCandidateHandler {
	return &ClipCandidateHandler{clipService: clipService}
}
func (h *ClipCandidateHandler) GetCandidates(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	candidates, err := h.clipService.GetCandidates(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to retrieve clip candidates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]any{
		"video_id":   id,
		"count":      len(candidates),
		"candidates": candidates,
	})
}
