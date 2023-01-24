package repositories

import (
	"context"
	"database/sql"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/google/uuid"
	"log"
)

type DatabaseRepository struct {
	Storage *sql.DB
}

func (repo *DatabaseRepository) Create(ctx context.Context, shortURL entities.ShortURL) (entities.ShortURL, error) {
	_, err := repo.Storage.ExecContext(
		ctx,
		"INSERT INTO short_urls (id, short_url, original_url, user_id) values ($1, $2, $3, $4)",
		shortURL.ID, shortURL.Short, shortURL.Original, shortURL.UserID.String(),
	)
	log.Printf(
		"Query to execute: INSERT INTO short_urls (id, short_url, original_url, user_id) values (%s, %s, %s, %s)",
		shortURL.ID, shortURL.Short, shortURL.Original, shortURL.UserID,
	)
	if err != nil {
		return entities.ShortURL{}, err
	}
	return shortURL, nil
}

func (repo *DatabaseRepository) GetByID(ctx context.Context, id string) (entities.ShortURL, bool) {
	var shortURL entities.ShortURL
	row := repo.Storage.QueryRowContext(
		ctx,
		"SELECT id, short_url, original_url, user_id FROM short_urls WHERE id = $1",
		id,
	)
	err := row.Scan(&shortURL.ID, &shortURL.Short, &shortURL.Original, &shortURL.UserID)
	if err != nil {
		return entities.ShortURL{}, false
	}
	return shortURL, true
}

func (repo *DatabaseRepository) GetByUserID(ctx context.Context, userID uuid.UUID) []entities.ShortURL {
	shortURLs := make([]entities.ShortURL, 0, 16)

	rows, err := repo.Storage.QueryContext(
		ctx,
		"SELECT id, short_url, original_url, user_id from short_urls WHERE user_id = $1",
		userID.String(),
	)
	if err != nil {
		return shortURLs
	}
	defer rows.Close()

	for rows.Next() {
		var shortURL entities.ShortURL
		err = rows.Scan(&shortURL.ID, &shortURL.Short, &shortURL.Original, &shortURL.UserID)
		if err != nil {
			log.Fatal(err)
		}

		shortURLs = append(shortURLs, shortURL)
	}

	err = rows.Err()
	if err != nil {
		log.Printf("Error happened while fetching urls: %v", err)
	}

	return shortURLs
}
