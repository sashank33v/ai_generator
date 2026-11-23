package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"generator/internal/service"

	"github.com/jackc/pgx/v5"
)

type ProcessingJobHandler struct {
	jobService *service.ProcessingJobService
}

func NewProcessingJobHandler(jobService *service.ProcessingJobService) *ProcessingJobHandler {
	return &ProcessingJobHandler{
		jobService: jobService,
	}
}

func (h *ProcessingJobHandler) GetJobByID(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil || id <= 0 {
		log.Printf("GJBI-1%v\n", err)
		http.Error(w, "invalid job ID", http.StatusBadRequest)
		return
	}

	job, err := h.jobService.GetJobByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("GJBI-2%v\n", err)
			http.Error(w, "processing job not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to retrieve processing job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(job); err != nil {
		return
	}
}
