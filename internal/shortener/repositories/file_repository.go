package repositories

import (
	"context"
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

func (repo *FileRepository) GetByID(ctx context.Context, id string) (entities.ShortURL, bool, error) {
	lock.RLock()
	result, exist := repo.Storage[id]
	lock.RUnlock()
	return result, exist, nil
}

func (repo *FileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entities.ShortURL, error) {
	result := make([]entities.ShortURL, 0, 8)
	lock.RLock()
	for _, shortURL := range repo.Storage {
		if shortURL.UserID == userID && shortURL.IsActive {
			result = append(result, shortURL)
		}
	}
	lock.RUnlock()
	return result, nil
}

func (repo *FileRepository) Create(ctx context.Context, shortURL entities.ShortURL) (entities.ShortURL, error) {
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

func (repo *FileRepository) CreateMultiple(
	ctx context.Context,
	urls []entities.ShortURL,
) ([]entities.ShortURL, error) {
	lock.Lock()
	for _, url := range urls {
		repo.Storage[url.ID] = url
	}
	lock.Unlock()
	file, err := repo.openStorageFile()
	if err != nil {
		return []entities.ShortURL{}, err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(&repo.Storage); err != nil {
		return []entities.ShortURL{}, err
	}
	return urls, nil
}

func (repo *FileRepository) DeleteRecords(ctx context.Context, userID uuid.UUID, ids []string) error {
	lock.Lock()
	for _, id := range ids {
		if url, exist := repo.Storage[id]; exist && url.UserID == userID {
			url.IsActive = false
			repo.Storage[id] = url
		}
	}
	lock.Unlock()

	file, err := repo.openStorageFile()
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err = encoder.Encode(&repo.Storage); err != nil {
		return err
	}

	return nil
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
