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
//
// Run with initial flags:
//
//	go run -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=initial commit'" cmd/shortener/main.go -a localhost:8080 -b http://localhost:8080 -f storage.json
package main

import (
	"log"
	"net/http"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/handlers"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/utils"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func main() {
	repo := utils.SetRepository()
	h := handlers.NewShortener(repo)
	log.Printf("Build version: %s", buildVersion)
	log.Printf("Build date: %s", buildDate)
	log.Printf("Build commit: %s", buildCommit)
	log.Fatal(http.ListenAndServe(config.Settings.ServerAddress, h))
}
