package routes

import (
	"net/http"

	"generator/internal/handler"
)

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthHandler)

	return mux
}
