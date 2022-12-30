package utils

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"log"
)

func SetRepositories() (repositories.Repository, repositories.Repository) {
	if config.Settings.FileStoragePath == "" {
		repo := repositories.InMemoryRepository{Storage: make(map[string]string)}
		backwardRepo := repositories.InMemoryRepository{Storage: make(map[string]string)}

		log.Println("In memory storage`s been chosen")
		return &repo, &backwardRepo
	}

	repo := repositories.FileRepository{
		Storage:  make(map[string]string),
		FilePath: config.Settings.FileStoragePath,
	}
	repo.Restore()

	backwardRepo := repositories.FileRepository{
		Storage:  make(map[string]string),
		FilePath: "back_" + config.Settings.FileStoragePath,
	}
	backwardRepo.Restore()

	log.Println("File storage`s been  chosen")
	return &repo, &backwardRepo
}
