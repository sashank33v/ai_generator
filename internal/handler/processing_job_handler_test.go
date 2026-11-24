package handler

import (
	"context"
	"errors"
	"generator/internal/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type mockProcessingJobService struct {
	job *model.ProcessingJob
	err error
}

func (m *mockProcessingJobService) GetJobByID(ctx context.Context, id int64) (*model.ProcessingJob, error) {
	return m.job, m.err
}
func TestGetJobByID(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		jobID          string
		mockJob        *model.ProcessingJob
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "successful job retrieval",
			jobID: "1",
			mockJob: &model.ProcessingJob{
				ID:        1,
				VideoID:   10,
				Status:    "COMPLETED",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"COMPLETED"`,
		},
		{
			name:           "invalid job ID",
			jobID:          "abc",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid job ID",
		},
		{
			name:           "job not found",
			jobID:          "999",
			mockErr:        pgx.ErrNoRows,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "processing job not found",
		},
		{
			name:           "internal service error",
			jobID:          "1",
			mockErr:        errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to retrieve processing job",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockProcessingJobService{
				job: tt.mockJob,
				err: tt.mockErr,
			}

			handler := NewProcessingJobHandler(mockService)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/jobs/"+tt.jobID,
				nil,
			)
			req.SetPathValue("id", tt.jobID)

			recorder := httptest.NewRecorder()

			handler.GetJobByID(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf(
					"expected status %d, got %d",
					tt.expectedStatus,
					recorder.Code,
				)
			}

			if !strings.Contains(recorder.Body.String(), tt.expectedBody) {
				t.Errorf(
					"expected body to contain %q, got %q",
					tt.expectedBody,
					recorder.Body.String(),
				)
			}
		})
	}
}
