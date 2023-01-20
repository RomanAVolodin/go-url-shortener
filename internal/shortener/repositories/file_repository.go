package repositories

import (
	"encoding/json"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
)

type FileRepository struct {
	Storage  map[string]entities.ShortUrl
	FilePath string
}

func (repo *FileRepository) GetByID(id string) (entities.ShortUrl, bool) {
	lock.RLock()
	result, exist := repo.Storage[id]
	lock.RUnlock()
	return result, exist
}

func (repo *FileRepository) GetByUserId(userId uuid.UUID) []entities.ShortUrl {
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

func (repo *FileRepository) Create(shortUrl entities.ShortUrl) (entities.ShortUrl, error) {
	lock.Lock()
	repo.Storage[shortUrl.Id] = shortUrl
	lock.Unlock()

	file, err := repo.openStorageFile()
	if err != nil {
		return entities.ShortUrl{}, err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(&repo.Storage); err != nil {
		return entities.ShortUrl{}, err
	}
	return shortUrl, nil
}

func (repo *FileRepository) Restore() error {
	file, err := repo.openStorageFile()
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&repo.Storage); err == io.EOF {
		log.Println("State restored")
	} else if err != nil {
		return err
	}
	return nil
}

func (repo *FileRepository) openStorageFile() (*os.File, error) {
	file, err := os.OpenFile(repo.FilePath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return file, nil
}
