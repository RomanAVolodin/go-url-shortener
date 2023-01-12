package utils

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"log"
)

func SetRepositories() (repositories.Repository, repositories.Repository) {
	if config.Settings.FileStoragePath != "" {
		repo := repositories.FileRepository{
			Storage:  make(map[string]string),
			FilePath: config.Settings.FileStoragePath,
		}

		backwardRepo := repositories.FileRepository{
			Storage:  make(map[string]string),
			FilePath: config.Settings.FileStoragePath + "_back",
		}

		if err, errBack := repo.Restore(), backwardRepo.Restore(); err == nil && errBack == nil {
			log.Println("File storage`s been  chosen")
			return &repo, &backwardRepo
		}
		log.Println("Error while choosing file storage")
	}

	repo := repositories.InMemoryRepository{Storage: make(map[string]string)}
	backwardRepo := repositories.InMemoryRepository{Storage: make(map[string]string)}

	log.Println("In memory storage`s been chosen")
	return &repo, &backwardRepo
}
