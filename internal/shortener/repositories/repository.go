// Package repositories declares IRepository interface and all of its implementations.
package repositories

import (
	"context"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/google/uuid"
)

// IRepository interface for ShortUrl repositories.
type IRepository interface {
	GetByID(ctx context.Context, id string) (entities.ShortURL, bool, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entities.ShortURL, error)
	Create(ctx context.Context, shortURL entities.ShortURL) (entities.ShortURL, error)
	CreateMultiple(ctx context.Context, urls []entities.ShortURL) ([]entities.ShortURL, error)
	DeleteRecords(ctx context.Context, userID uuid.UUID, ids []string) error
}
