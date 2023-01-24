package repositories

import (
	"context"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/google/uuid"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entities.ShortURL, bool)
	GetByUserID(ctx context.Context, userID uuid.UUID) []entities.ShortURL
	Create(ctx context.Context, shortURL entities.ShortURL) (entities.ShortURL, error)
}
