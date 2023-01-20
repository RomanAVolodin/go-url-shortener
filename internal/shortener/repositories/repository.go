package repositories

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/google/uuid"
)

type Repository interface {
	GetByID(id string) (entities.ShortURL, bool)
	GetByUserID(userID uuid.UUID) []entities.ShortURL
	Create(shortURL entities.ShortURL) (entities.ShortURL, error)
}
