package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/tommyhedley/fiberytsheets/internal/handlers"
	"github.com/tommyhedley/fiberytsheets/internal/handlers/oauth2"
	"github.com/tommyhedley/fiberytsheets/internal/handlers/synchronizer"
)

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handlers.Config)
	mux.HandleFunc("GET /logo", handlers.Logo)

	mux.HandleFunc("POST /oauth2/v1/authorize", oauth2.AuthorizeHandler)
	mux.HandleFunc("POST /oauth2/v1/access_token", oauth2.TokenHandler)
	mux.HandleFunc("POST /validate", oauth2.ValidateHandler)

	mux.HandleFunc("POST /api/v1/synchronizer/config", synchronizer.Config)
	mux.HandleFunc("POST /api/v1/synchronizer/schema", synchronizer.Schema)
	mux.HandleFunc("POST /api/v1/synchronizer/filter/validate", synchronizer.ValidateFilters)
	mux.HandleFunc("POST /api/v1/synchronizer/data", synchronizer.Data)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server started on port: %s", port)
	log.Fatal(srv.ListenAndServe())
}
