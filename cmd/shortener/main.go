package main

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/handlers"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"log"
	"net/http"
)

func main() {

	repo := make(repositories.UrlsRepository)
	backwardRepo := make(repositories.UrlsRepository)

	h := handlers.NewShortenerHandler(repo, backwardRepo)
	log.Printf("Server started at %s", config.Settings.ServerAddress)
	log.Fatal(http.ListenAndServe(config.Settings.ServerAddress, h))
}
