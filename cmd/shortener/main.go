// Package main runs the application.
// Main ingress point in entire application.
// Starts WEB server.
//
// # Application may be started with several ways
//
// With explicit environment variables
//
//	SERVER_ADDRESS=localhost:8080 BASE_URL=http://localhost:8080 go run cmd/shortener/main.go
//
// With flags to set host and port. In memory storage:
//
//	go run cmd/shortener/main.go -a localhost:8080 -b http://localhost:8080
//
// File storage:
//
//	go run cmd/shortener/main.go -a localhost:8080 -b http://localhost:8080 -f storage.json
//
// Database storage:
//
//	go run cmd/shortener/main.go -a localhost:8080 -b http://localhost:8080 -f storage.json -d postgres://shortener:secret@localhost:5432/shortener
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/handlers"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/utils"
)

func main() {
	repo := utils.SetRepository()
	h := handlers.NewShortener(repo)
	log.Printf("Server started at %s", config.Settings.ServerAddress)
	log.Fatal(http.ListenAndServe(config.Settings.ServerAddress, h))
	os.Exit(0)
}
