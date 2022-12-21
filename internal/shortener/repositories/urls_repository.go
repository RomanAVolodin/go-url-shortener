package repositories

import (
	"github.com/lithammer/shortuuid"
	"sync"
)

type UrlsRepository map[string]string

var lock = sync.RWMutex{}

func (repo UrlsRepository) GetByID(id string) (string, bool) {
	lock.RLock()
	result, exist := repo[id]
	lock.RUnlock()
	return result, exist
}

func (repo UrlsRepository) Create(url string) string {
	shortURL := shortuuid.New()
	lock.Lock()
	repo[shortURL] = url
	lock.Unlock()
	return shortURL
}

func (repo UrlsRepository) CreateBackwardRecord(url, id string) {
	lock.Lock()
	repo[url] = id
	lock.Unlock()
}
