package handler

import (
	"encoding/json"
	"net/http"

	"generator/internal/service"
)

type VideoHandler struct {
	videoService *service.VideoService
}

func NewVideoHandler(videoService *service.VideoService) *VideoHandler {
	return &VideoHandler{
		videoService: videoService,
	}
}

func (h *VideoHandler) UploadVideo(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 500 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+(1<<20))

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Invalid upload request or file too large", http.StatusBadRequest)
		return
	}

	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Video file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	video, err := h.videoService.UploadVideo(r.Context(), file, header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Video uploaded successfully",
		"video_id": video.ID,
		"status":   video.Status,
	})
}
