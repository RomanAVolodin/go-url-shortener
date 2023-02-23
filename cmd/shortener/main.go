package main

import (
	"log"
	"net/http"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/handlers"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/utils"
)

func main() {
	repo := utils.SetRepository()
	h := handlers.NewShortenerHandler(repo)
	log.Printf("Server started at %s", config.Settings.ServerAddress)
	log.Fatal(http.ListenAndServe(config.Settings.ServerAddress, h))
}
