package repositories

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/google/uuid"
)

type Repository interface {
	GetByID(id string) (entities.ShortUrl, bool)
	GetByUserId(userId uuid.UUID) []entities.ShortUrl
	Create(shortUrl entities.ShortUrl) (entities.ShortUrl, error)
}
