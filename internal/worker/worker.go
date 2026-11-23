package worker

import (
	"context"
	"log"
	"time"

	"generator/internal/service"
)

type Worker struct {
	workerService *service.WorkerService
	pollInterval  time.Duration
}

func NewWorker(
	workerService *service.WorkerService,
) *Worker {
	return &Worker{
		workerService: workerService,
		pollInterval:  5 * time.Second,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	log.Println("Background worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Background worker stopped")
			return

		case <-ticker.C:
			job, err := w.workerService.ProcessNextJob(ctx)
			if err != nil {
				log.Printf("Failed to claim job: %v", err)
				continue
			}

			if job == nil {
				continue
			}
			log.Printf("Started processing job ID: %d", job.ID)

			err = w.workerService.ProcessJob(ctx, job)

			if err != nil {
				log.Printf("Failed to process job %d: %v", job.ID, err)

				errorMessage := err.Error()

				statusErr := w.workerService.UpdateJobStatus(
					ctx,
					job,
					"FAILED",
					&errorMessage,
				)

				if statusErr != nil {
					log.Printf("Failed to update job status: %v", statusErr)
				}

				continue
			}

			err = w.workerService.UpdateJobStatus(
				ctx,
				job,
				"COMPLETED",
				nil,
			)

			if err != nil {
				log.Printf("Failed to mark job as completed: %v", err)
				continue
			}

			log.Printf("Job %d completed successfully", job.ID)

		}
	}
}
