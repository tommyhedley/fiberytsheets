package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/tommyhedley/fiberytsheets/internal/handlers"
)

func main() {
	godotenv.Load()
	config := handlers.AppConfig{
		ID:          "qbtime",
		Name:        "Quickbooks Time",
		Version:     "0.1.0",
		Description: "Integrate Quickbooks Time with Fibery",
		Authentication: []struct {
			ID          string   `json:"id"`
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Fields      []string `json:"fields"`
		}{
			{
				ID:          "public",
				Name:        "Public Access",
				Description: "There is no authentication required",
				Fields:      []string{},
			},
		},
		Sources: []string{},
		ResponsibleFor: struct {
			DataSynchronization bool `json:"dataSynchronization"`
		}{
			DataSynchronization: true,
		},
	}

	port := os.Getenv("PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", handlers.Readiness)
	mux.HandleFunc("GET /", config.Config)
	mux.HandleFunc("POST /oauth2/v1/authorize", handlers.Authorize)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server started on port: %s", port)
	log.Fatal(srv.ListenAndServe())
}
