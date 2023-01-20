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
	Storage  map[string]entities.ShortURL
	FilePath string
}

func (repo *FileRepository) GetByID(id string) (entities.ShortURL, bool) {
	lock.RLock()
	result, exist := repo.Storage[id]
	lock.RUnlock()
	return result, exist
}

func (repo *FileRepository) GetByUserID(userID uuid.UUID) []entities.ShortURL {
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

func (repo *FileRepository) Create(shortURL entities.ShortURL) (entities.ShortURL, error) {
	lock.Lock()
	repo.Storage[shortURL.ID] = shortURL
	lock.Unlock()

	file, err := repo.openStorageFile()
	if err != nil {
		return entities.ShortURL{}, err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(&repo.Storage); err != nil {
		return entities.ShortURL{}, err
	}
	return shortURL, nil
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
