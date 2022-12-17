package repositories

import "github.com/lithammer/shortuuid"

type UrlsRepository map[string]string

func (repo UrlsRepository) GetById(id string) (string, bool) {
	result, exist := repo[id]
	return result, exist
}

func (repo UrlsRepository) Create(url string) string {
	shortURL := shortuuid.New()
	repo[shortURL] = url
	return shortURL
}

func (repo UrlsRepository) CreateBackwardRecord(url, id string) {
	repo[url] = id
}
