package utils

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"log"
)

func SetRepository() repositories.Repository {
	if config.Settings.FileStoragePath != "" {
		repo := repositories.FileRepository{
			Storage:  make(map[string]entities.ShortURL),
			FilePath: config.Settings.FileStoragePath,
		}

		if err := repo.Restore(); err == nil {
			log.Println("File storage`s been  chosen")
			return &repo
		}
		log.Println("Error while choosing file storage")
	}

	repo := repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)}

	log.Println("In memory storage`s been chosen")
	return &repo
}
