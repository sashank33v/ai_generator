package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"generator/internal/ai"
	"generator/internal/config"
	"generator/internal/handler"
	"generator/internal/processor"
	"generator/internal/repository"
	"generator/internal/routes"
	"generator/internal/service"
	"generator/internal/worker"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, err := repository.NewPostgresPool(context.Background(), cfg)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	llamaClient := ai.NewLlamaClient()

	videoRepo := repository.NewVideoRepository(db)
	jobRepo := repository.NewProcessingJobRepository(db)
	workerRepo := repository.NewWorkerRepository(db)
	generatedClipRepo := repository.NewGeneratedClipRepository(db)
	audioProcessor := processor.NewAudioProcessor()
	transcriptionProcessor := processor.NewTranscriptionProcessor()
	videoProcessor := processor.NewVideoProcessor()
	transcriptionRepo := repository.NewTranscriptionRepository(db)
	jobService := service.NewProcessingJobService(jobRepo)
	clipService := service.NewClipCandidateService(transcriptionRepo)
	videoService := service.NewVideoService(videoRepo, jobService, cfg.StorageDir)
	scoringService := service.NewClipScoringService(llamaClient)
	rankingService := service.NewClipRankingService(clipService, scoringService)
	metadataService := service.NewClipMetadataService(llamaClient)
	generationService := service.NewClipGenerationService(rankingService, videoRepo, videoProcessor, metadataService, generatedClipRepo)
	workerService := service.NewWorkerService(workerRepo, videoRepo, audioProcessor, transcriptionProcessor, transcriptionRepo, generationService)
	videoHandler := handler.NewVideoHandler(videoService)
	jobHandler := handler.NewProcessingJobHandler(jobService)
	rankingHandler := handler.NewClipRankingHandler(rankingService)
	generationHandler := handler.NewClipGenerationHandler(generationService)
	clipHandler := handler.NewClipCandidateHandler(clipService)
	retryHandler := handler.NewRetryJobHandler(workerService)
	router := routes.SetupRoutes(videoHandler, jobHandler, clipHandler, rankingHandler, generationHandler, retryHandler)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()
	recoveryCtx, recoveryCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	err = workerService.RecoverInterruptedJobs(recoveryCtx)
	recoveryCancel()

	if err != nil {
		log.Fatalf("Failed to recover interrupted jobs: %v", err)
	}

	log.Println("Interrupted job recovery completed")

	backgroundWorker := worker.NewWorker(workerService)
	go backgroundWorker.Start(ctx)

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("%s running on port %s", cfg.AppName, cfg.Port)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received")

	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	stop()

	log.Println("Application stopped gracefully")
}
