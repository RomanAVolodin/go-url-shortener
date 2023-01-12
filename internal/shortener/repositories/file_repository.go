package repositories

import (
	"encoding/json"
	"github.com/lithammer/shortuuid"
	"io"
	"log"
	"os"
)

type FileRepository struct {
	Storage  map[string]string
	FilePath string
}

func (repo *FileRepository) GetByID(id string) (string, bool) {
	lock.RLock()
	result, exist := repo.Storage[id]
	lock.RUnlock()
	return result, exist
}

func (repo *FileRepository) CreateSave(url string) (string, error) {
	shortURL := shortuuid.New()
	return repo.Save(shortURL, url)
}

func (repo *FileRepository) Save(key, value string) (string, error) {
	lock.Lock()
	repo.Storage[key] = value
	lock.Unlock()

	file, err := repo.openStorageFile()
	if err != nil {
		return "", err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(&repo.Storage); err != nil {
		return "", err
	}
	return key, nil
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
