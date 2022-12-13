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

	http.HandleFunc("/", handlers.ShortenerHandler(repo, backwardRepo))
	log.Fatal(http.ListenAndServe(config.AppPort, nil))
}
