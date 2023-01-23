package repositories

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/google/uuid"
	"sync"
)

type InMemoryRepository struct {
	Storage map[string]entities.ShortURL
}

var lock = sync.RWMutex{}

func (repo *InMemoryRepository) GetByID(id string) (entities.ShortURL, bool) {
	lock.RLock()
	result, exist := repo.Storage[id]
	lock.RUnlock()
	return result, exist
}

func (repo *InMemoryRepository) GetByUserID(userID uuid.UUID) []entities.ShortURL {
	result := make([]entities.ShortURL, 0, 8)
	lock.RLock()
	for _, shortURL := range repo.Storage {
		if shortURL.UserID == userID {
			result = append(result, shortURL)
		}
	}
	lock.RUnlock()
	return result
}

func (repo *InMemoryRepository) Create(shortURL entities.ShortURL) (entities.ShortURL, error) {
	lock.Lock()
	repo.Storage[shortURL.ID] = shortURL
	lock.Unlock()
	return shortURL, nil
}
