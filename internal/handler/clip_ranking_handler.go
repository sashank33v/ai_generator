package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"generator/internal/service"
)

type ClipRankingHandler struct {
	rankingService *service.ClipRankingService
}

func NewClipRankingHandler(
	rankingService *service.ClipRankingService,
) *ClipRankingHandler {
	return &ClipRankingHandler{
		rankingService: rankingService,
	}
}

func (h *ClipRankingHandler) GetRankings(
	w http.ResponseWriter,
	r *http.Request,
) {
	videoID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || videoID <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	rankings, err := h.rankingService.RankCandidates(
		r.Context(),
		videoID,
	)
	if err != nil {
		log.Printf("clip ranking failed for video %d: %v", videoID, err)
		http.Error(w, "failed to rank clips", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"video_id":  videoID,
		"count":     len(rankings),
		"best_clip": rankings[0],
		"rankings":  rankings,
	}); err != nil {
		log.Printf("failed to encode rankings: %v", err)
	}
}
