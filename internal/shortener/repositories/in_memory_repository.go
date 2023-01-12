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

func (repo *InMemoryRepository) CreateSave(url string) (string, error) {
	shortURL := shortuuid.New()
	return repo.Save(shortURL, url)
}

func (repo *InMemoryRepository) Save(key, value string) (string, error) {
	lock.Lock()
	repo.Storage[key] = value
	lock.Unlock()
	return key, nil
}
