package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"generator/internal/repository"
	"generator/internal/service"

	"github.com/jackc/pgx/v5"
)

type RetryJobHandler struct {
	workerService *service.WorkerService
}

func NewRetryJobHandler(
	workerService *service.WorkerService,
) *RetryJobHandler {
	return &RetryJobHandler{
		workerService: workerService,
	}
}

func (h *RetryJobHandler) RetryJob(
	w http.ResponseWriter,
	r *http.Request,
) {

	jobID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil || jobID <= 0 {
		http.Error(w, "invalid job ID", http.StatusBadRequest)
		return
	}

	err = h.workerService.RetryJob(r.Context(), jobID)

	if err != nil {

		switch {
		case errors.Is(err, pgx.ErrNoRows):
			http.Error(w, "job not found", http.StatusNotFound)

		case errors.Is(err, repository.ErrJobNotRetryable):
			http.Error(w, err.Error(), http.StatusConflict)

		default:
			log.Printf("Job retry failed: %v", err)
			http.Error(w, "failed to retry job", http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "Job queued for retry",
		"job_id":  jobID,
		"status":  "PENDING",
	})
}
