package repositories

import (
	"github.com/lithammer/shortuuid"
	"sync"
)

type InMemoryRepository struct {
	Storage map[string]string
}

var lock = sync.RWMutex{}

func (repo *InMemoryRepository) GetByID(id string) (string, bool) {
	lock.RLock()
	result, exist := repo.Storage[id]
	lock.RUnlock()
	return result, exist
}

func (repo *InMemoryRepository) CreateSave(url string) string {
	shortURL := shortuuid.New()
	repo.Save(shortURL, url)
	return shortURL
}

func (repo *InMemoryRepository) Save(key, value string) {
	lock.Lock()
	repo.Storage[key] = value
	lock.Unlock()
}
