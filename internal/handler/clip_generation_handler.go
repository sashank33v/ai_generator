package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"generator/internal/service"

	"github.com/jackc/pgx/v5"
)

type ClipGenerationHandler struct {
	generationService *service.ClipGenerationService
}

func NewClipGenerationHandler(generationService *service.ClipGenerationService) *ClipGenerationHandler {
	return &ClipGenerationHandler{
		generationService: generationService,
	}
}

func (h *ClipGenerationHandler) GenerateBestClip(w http.ResponseWriter, r *http.Request) {
	videoID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || videoID <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	outputPath, bestClip, metadata, err := h.generationService.GenerateBestClip(r.Context(), videoID, nil)
	if err != nil {
		log.Printf("clip generation failed for video %d: %v", videoID, err)
		http.Error(w, "failed to generate clip", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(map[string]any{
		"message":     "Clip generated successfully",
		"video_id":    videoID,
		"output_path": outputPath,
		"best_clip":   bestClip,
		"metadata":    metadata,
	}); err != nil {
		log.Printf("failed to encode generated clip response: %v", err)
	}
}
func (h *ClipGenerationHandler) DownloadClip(w http.ResponseWriter, r *http.Request) {
	videoID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || videoID <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	filePath, err := h.generationService.GetGeneratedClipPath(
		r.Context(),
		videoID,
	)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) ||
			errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "generated clip not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to retrieve clip", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=%q", filepath.Base(filePath)),
	)

	http.ServeFile(w, r, filePath)
}

func (h *ClipGenerationHandler) GetGeneratedClip(w http.ResponseWriter, r *http.Request) {
	videoID, err := strconv.ParseInt(
		r.PathValue("id"), 10, 64,
	)

	if err != nil || videoID <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	clip, err := h.generationService.GetGeneratedClip(
		r.Context(),
		videoID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "generated clip not found", http.StatusNotFound)
			return
		}

		log.Printf("failed to retrieve generated clip: %v", err)
		http.Error(w, "failed to retrieve clip", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(clip); err != nil {
		log.Printf("failed to encode clip: %v", err)
	}
}
