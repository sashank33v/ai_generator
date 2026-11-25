package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"generator/internal/downloader"
	"generator/internal/service"
)

type YouTubeHandler struct {
	youtubeService *service.YouTubeService
}

type YouTubeRequest struct {
	URL string `json:"url"`
}

func NewYouTubeHandler(youtubeService *service.YouTubeService) *YouTubeHandler {
	return &YouTubeHandler{
		youtubeService: youtubeService,
	}
}

func (h *YouTubeHandler) ProcessYouTubeURL(w http.ResponseWriter, r *http.Request) {

	var request YouTubeRequest

	// Limit incoming JSON body to 4 KB.
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := decoder.Decode(new(any)); err != io.EOF {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Allow sufficient time for downloading the source video.
	controller := http.NewResponseController(w)

	if err := controller.SetWriteDeadline(time.Now().Add(11 * time.Minute)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		log.Printf("Failed to set response deadline: %v", err)
		http.Error(w, "failed to prepare request", http.StatusInternalServerError)
		return
	}

	// Process the YouTube URL.
	video, job, err := h.youtubeService.ProcessYouTubeURL(r.Context(), request.URL)

	if err != nil {
		switch {
		case errors.Is(err, downloader.ErrInvalidYouTubeURL):

			http.Error(w, "invalid YouTube video URL", http.StatusBadRequest)

		case errors.Is(err, service.ErrYouTubeDownload):

			log.Printf("YouTube download failed: %v", err)

			http.Error(w, "failed to download YouTube video", http.StatusBadGateway)

		default:

			log.Printf("YouTube processing failed: %v", err)

			http.Error(w, "failed to process YouTube video", http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(map[string]any{
		"message":  "YouTube video downloaded and queued",
		"video_id": video.ID,
		"job_id":   job.ID,
		"status":   job.Status,
	})
}
