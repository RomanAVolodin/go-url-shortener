package main

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/handlers"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/utils"
	"log"
	"net/http"
)

func main() {
	repo, backRepo := utils.SetRepositories()
	h := handlers.NewShortenerHandler(repo, backRepo)
	log.Printf("Server started at %s", config.Settings.ServerAddress)
	log.Fatal(http.ListenAndServe(config.Settings.ServerAddress, h))
}
