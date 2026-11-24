package routes

import (
	"net/http"

	"generator/internal/handler"
)

func SetupRoutes(videoHandler *handler.VideoHandler, jobHandler *handler.ProcessingJobHandler, clipHandler *handler.ClipCandidateHandler, rankingHandler *handler.ClipRankingHandler, generationHandler *handler.ClipGenerationHandler, retryHandler *handler.RetryJobHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/videos/upload", videoHandler.UploadVideo)
	mux.HandleFunc("GET /api/jobs/{id}", jobHandler.GetJobByID)

	mux.HandleFunc("GET /api/videos/{id}/candidates", clipHandler.GetCandidates)
	mux.HandleFunc("GET /api/videos/{id}/rankings", rankingHandler.GetRankings)
	mux.HandleFunc("POST /api/videos/{id}/generate", generationHandler.GenerateBestClip)
	mux.HandleFunc("GET /api/videos/{id}/download", generationHandler.DownloadClip)
	mux.HandleFunc("GET /api/videos/{id}/clip", generationHandler.GetGeneratedClip)
	mux.HandleFunc("POST /api/jobs/{id}/retry", retryHandler.RetryJob)
	return mux
}
