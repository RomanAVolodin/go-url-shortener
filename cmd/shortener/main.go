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
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"

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

	manager := &autocert.Manager{
		Cache:      autocert.DirCache("cache-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("my.domain.ru"),
	}
	server := &http.Server{
		Addr:      ":443",
		Handler:   h,
		TLSConfig: manager.TLSConfig(),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		log.Printf("Build version: %s", buildVersion)
		log.Printf("Build date: %s", buildDate)
		log.Printf("Build commit: %s", buildCommit)
		if config.Settings.EnableHTTPS {
			log.Fatal(server.ListenAndServeTLS("", ""))
		} else {
			log.Fatal(http.ListenAndServe(config.Settings.ServerAddress, h))
		}
	}()

	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed with error:%+v", err)
	}
	log.Print("Server Exited Properly")
}
