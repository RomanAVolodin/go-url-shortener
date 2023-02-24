package repositories

import (
	"context"
	"sync"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/google/uuid"
)

// InMemoryRepository repository based memory storage.
type InMemoryRepository struct {
	Storage  map[string]entities.ShortURL
	ToDelete chan entities.ItemToDelete
}

var lock = sync.RWMutex{}

func (repo *InMemoryRepository) GetByID(ctx context.Context, id string) (entities.ShortURL, bool, error) {
	lock.RLock()
	result, exist := repo.Storage[id]
	lock.RUnlock()
	return result, exist, nil
}

func (repo *InMemoryRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entities.ShortURL, error) {
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

func (repo *InMemoryRepository) Create(ctx context.Context, shortURL entities.ShortURL) (entities.ShortURL, error) {
	lock.Lock()
	repo.Storage[shortURL.ID] = shortURL
	lock.Unlock()
	return shortURL, nil
}

func (repo *InMemoryRepository) CreateMultiple(
	ctx context.Context,
	urls []entities.ShortURL,
) ([]entities.ShortURL, error) {
	lock.Lock()
	for _, url := range urls {
		repo.Storage[url.ID] = url
	}
	lock.Unlock()
	return urls, nil
}

func (repo *InMemoryRepository) DeleteRecords(ctx context.Context, userID uuid.UUID, ids []string) error {
	lock.Lock()
	for _, id := range ids {
		if url, exist := repo.Storage[id]; exist && url.UserID == userID {
			url.IsActive = false
			repo.Storage[id] = url
		}
	}
	lock.Unlock()
	return nil
}
