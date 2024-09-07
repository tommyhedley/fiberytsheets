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

	port := os.Getenv("PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", handlers.Readiness)
	mux.HandleFunc("GET /", handlers.Config)
	mux.HandleFunc("POST /oauth2/v1/authorize", handlers.Authorize)
	mux.HandleFunc("POST /oauth2/v1/access_token", handlers.GetToken)
	mux.HandleFunc("POST /validate", handlers.ValidateRefresh)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server started on port: %s", port)
	log.Fatal(srv.ListenAndServe())
}
