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
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/RomanAVolodin/go-url-shortener/internal/shortener/proto"
	"google.golang.org/grpc"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/handlers/grpcserver"

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

	handler := handlers.NewShortener(repo)

	server := &http.Server{
		Addr:    config.Settings.ServerAddress,
		Handler: handler,
	}

	listen, err := net.Listen("tcp", config.Settings.GrpcPort)
	if err != nil {
		log.Fatal(err)
	}
	gRPCServer := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.UnaryUserIDInterceptor))

	pb.RegisterShortenerServer(gRPCServer, &grpcserver.ShortenerGrpc{Shortener: &handlers.Shortener{Repo: repo}})
	// starts instance of gRPC server
	go func() {
		log.Println("Server gRPC has been started")
		if err := gRPCServer.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}()

	idleConnectionsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		log.Println("Start shutting down process")

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Server Shutdown error: %v", err)
		}
		log.Println("Server HTTP is stopped")
		gRPCServer.GracefulStop()
		log.Println("Server gRPC is stopped")

		if err := handler.Repo.CloseConnection(); err != nil {
			log.Printf("Servers Repos closing error: %v", err)
		}

		close(idleConnectionsClosed)
	}()

	log.Printf("Build version: %s", buildVersion)
	log.Printf("Build date: %s", buildDate)
	log.Printf("Build commit: %s", buildCommit)
	log.Printf("Trusted subnet: %s", config.Settings.TrustedSubnet)
	if config.Settings.EnableHTTPS {
		fmt.Println("Starting HTTPS server")
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("my.domain.ru"),
		}
		server.Addr = ":4443"
		server.TLSConfig = manager.TLSConfig()
		if err := server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			log.Fatalf("HTTPs server ListenAndServeTLS Error: %v", err)
		}
	} else {
		fmt.Println("Starting insecure server")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe Error: %v", err)
		}
	}

	<-idleConnectionsClosed
	log.Printf("Server was closed gracefully!")
}
