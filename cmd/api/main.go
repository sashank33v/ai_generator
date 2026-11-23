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

	videoRepo := repository.NewVideoRepository(db)
	jobRepo := repository.NewProcessingJobRepository(db)
	workerRepo := repository.NewWorkerRepository(db)
	audioProcessor := processor.NewAudioProcessor()
	jobService := service.NewProcessingJobService(jobRepo)
	videoService := service.NewVideoService(videoRepo, jobService, cfg.StorageDir)
	workerService := service.NewWorkerService(workerRepo, videoRepo, audioProcessor)
	videoHandler := handler.NewVideoHandler(videoService)
	jobHandler := handler.NewProcessingJobHandler(jobService)

	router := routes.SetupRoutes(videoHandler, jobHandler)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

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
