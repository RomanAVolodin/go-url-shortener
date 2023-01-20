package repositories

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/google/uuid"
	"sync"
)

type InMemoryRepository struct {
	Storage map[string]entities.ShortUrl
}

var lock = sync.RWMutex{}

func (repo *InMemoryRepository) GetByID(id string) (entities.ShortUrl, bool) {
	lock.RLock()
	result, exist := repo.Storage[id]
	lock.RUnlock()
	return result, exist
}

func (repo *InMemoryRepository) GetByUserId(userId uuid.UUID) []entities.ShortUrl {
	result := make([]entities.ShortUrl, 0, 8)
	lock.RLock()
	for _, shortUrl := range repo.Storage {
		if shortUrl.UserId == userId {
			result = append(result, shortUrl)
		}
	}
	lock.RUnlock()
	return result
}

func (repo *InMemoryRepository) Create(shortUrl entities.ShortUrl) (entities.ShortUrl, error) {
	lock.Lock()
	repo.Storage[shortUrl.Id] = shortUrl
	lock.Unlock()
	return shortUrl, nil
}
