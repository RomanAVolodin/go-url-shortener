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

func (repo *FileRepository) CreateSave(url string) string {
	shortURL := shortuuid.New()
	repo.Save(shortURL, url)
	return shortURL
}

func (repo *FileRepository) Save(key, value string) {
	lock.Lock()
	repo.Storage[key] = value
	lock.Unlock()

	file := repo.openStorageFile()
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(&repo.Storage); err != nil {
		log.Fatal(err)
	}
}

func (repo *FileRepository) Restore() {
	file := repo.openStorageFile()
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&repo.Storage); err == io.EOF {
		log.Println("State restored")
	} else if err != nil {
		log.Fatal(err)
	}
}

func (repo *FileRepository) openStorageFile() *os.File {
	file, err := os.OpenFile(repo.FilePath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err)
	}
	return file
}