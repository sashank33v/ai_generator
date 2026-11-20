package main

import (
	"fmt"
	"generator/internal/routes"
	"net/http"
)

func main() {
	router := routes.SetupRoutes()

	fmt.Println("Generator server running on :8080")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
