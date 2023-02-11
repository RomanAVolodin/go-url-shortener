package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/shortenerrors"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"log"
)

type DatabaseRepository struct {
	Storage *sql.DB
}

func (repo *DatabaseRepository) Create(ctx context.Context, shortURL entities.ShortURL) (entities.ShortURL, error) {
	_, err := repo.Storage.ExecContext(
		ctx,
		"INSERT INTO short_urls (id, short_url, original_url, user_id, correlation_id) values ($1, $2, $3, $4, $5);",
		shortURL.ID, shortURL.Short, shortURL.Original, shortURL.UserID.String(), shortURL.CorrelationID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var existed entities.ShortURL
			row := repo.Storage.QueryRowContext(
				ctx,
				"SELECT id, short_url, original_url, user_id, correlation_id, is_active FROM short_urls WHERE original_url = $1;",
				shortURL.Original,
			)
			errExisted := row.Scan(
				&existed.ID,
				&existed.Short,
				&existed.Original,
				&existed.UserID,
				&existed.CorrelationID,
				&existed.IsActive,
			)
			if errExisted != nil {
				return entities.ShortURL{}, errExisted
			}
			return existed, shortenerrors.ErrItemAlreadyExists
		}
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
		"INSERT INTO short_urls (id, short_url, original_url, user_id, correlation_id) values ($1, $2, $3, $4, $5);",
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

func (repo *DatabaseRepository) GetByID(ctx context.Context, id string) (entities.ShortURL, bool, error) {
	var shortURL entities.ShortURL
	row := repo.Storage.QueryRowContext(
		ctx,
		"SELECT id, short_url, original_url, user_id, correlation_id, is_active FROM short_urls WHERE id = $1;",
		id,
	)
	err := row.Scan(
		&shortURL.ID,
		&shortURL.Short,
		&shortURL.Original,
		&shortURL.UserID,
		&shortURL.CorrelationID,
		&shortURL.IsActive,
	)
	if err != nil {
		return entities.ShortURL{}, false, err
	}
	return shortURL, true, nil
}

func (repo *DatabaseRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entities.ShortURL, error) {
	shortURLs := make([]entities.ShortURL, 0, 16)

	rows, err := repo.Storage.QueryContext(
		ctx,
		"SELECT id, short_url, original_url, user_id, correlation_id, is_active FROM short_urls WHERE is_active=true AND user_id = $1;",
		userID.String(),
	)
	if err != nil {
		return shortURLs, err
	}
	defer rows.Close()

	for rows.Next() {
		var shortURL entities.ShortURL
		err = rows.Scan(
			&shortURL.ID,
			&shortURL.Short,
			&shortURL.Original,
			&shortURL.UserID,
			&shortURL.CorrelationID,
			&shortURL.IsActive,
		)
		if err != nil {
			log.Fatal(err)
		}

		shortURLs = append(shortURLs, shortURL)
	}

	err = rows.Err()
	if err != nil {
		return shortURLs, err
	}

	return shortURLs, nil
}

func (repo *DatabaseRepository) DeleteRecords(ctx context.Context, userID uuid.UUID, ids []string) error {
	query, args, err := sqlx.In(
		"UPDATE short_urls SET is_active=false WHERE user_id = ? AND id IN (?)",
		userID.String(),
		ids,
	)
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	_, err = repo.Storage.ExecContext(
		ctx,
		query,
		args...,
	)
	return err
}
