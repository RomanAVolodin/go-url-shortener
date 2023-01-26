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
		"INSERT INTO short_urls (id, short_url, original_url, user_id, correlation_id) values ($1, $2, $3, $4, $5)",
		shortURL.ID, shortURL.Short, shortURL.Original, shortURL.UserID.String(), shortURL.CorrelationID,
	)
	if err != nil {
		return entities.ShortURL{}, err
	}
	return shortURL, nil
}

func (repo *DatabaseRepository) CreateMultiple(
	ctx context.Context,
	urls []entities.ShortURL,
) ([]entities.ShortURL, error) {
	tx, err := repo.Storage.Begin()
	if err != nil {
		return []entities.ShortURL{}, err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(
		ctx,
		"INSERT INTO short_urls (id, short_url, original_url, user_id, correlation_id) values ($1, $2, $3, $4, $5)",
	)
	if err != nil {
		return []entities.ShortURL{}, err
	}
	defer stmt.Close()

	for _, shortURL := range urls {
		if _, err = stmt.ExecContext(
			ctx,
			shortURL.ID,
			shortURL.Short,
			shortURL.Original,
			shortURL.UserID.String(),
			shortURL.CorrelationID,
		); err != nil {
			return []entities.ShortURL{}, err
		}
	}

	return urls, tx.Commit()
}

func (repo *DatabaseRepository) GetByID(ctx context.Context, id string) (entities.ShortURL, bool) {
	var shortURL entities.ShortURL
	row := repo.Storage.QueryRowContext(
		ctx,
		"SELECT id, short_url, original_url, user_id, correlation_id FROM short_urls WHERE id = @id",
		sql.Named("id", id),
	)
	err := row.Scan(&shortURL.ID, &shortURL.Short, &shortURL.Original, &shortURL.UserID, &shortURL.CorrelationID)
	if err != nil {
		return entities.ShortURL{}, false
	}
	return shortURL, true
}

func (repo *DatabaseRepository) GetByUserID(ctx context.Context, userID uuid.UUID) []entities.ShortURL {
	shortURLs := make([]entities.ShortURL, 0, 16)

	rows, err := repo.Storage.QueryContext(
		ctx,
		"SELECT id, short_url, original_url, user_id, correlation_id FROM short_urls WHERE user_id = $1",
		userID.String(),
	)
	if err != nil {
		return shortURLs
	}
	defer rows.Close()

	for rows.Next() {
		var shortURL entities.ShortURL
		err = rows.Scan(&shortURL.ID, &shortURL.Short, &shortURL.Original, &shortURL.UserID, &shortURL.CorrelationID)
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
