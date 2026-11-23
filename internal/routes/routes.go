package routes

import (
	"net/http"

	"generator/internal/handler"
)

func SetupRoutes(videoHandler *handler.VideoHandler, jobHandler *handler.ProcessingJobHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/videos/upload", videoHandler.UploadVideo)
	mux.HandleFunc("GET /api/jobs/{id}", jobHandler.GetJobByID)
	return mux
}
