package repositories

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/shortenerrors"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

// DatabaseRepository repository based on database.
type DatabaseRepository struct {
	Storage                    *sql.DB
	ToDelete                   chan *entities.ItemToDelete
	DeleteAccumulatorWaitGroup *sync.WaitGroup
}

// toDeleteWaitGroup waits for all toDelete goroutines
var toDeleteWaitGroup sync.WaitGroup

// lockURLToDeleteStorage mutex for deletion process.
var lockURLToDeleteStorage = sync.Mutex{}

// accumulateRecordsToDeleteStopper stops accumulator.
var accumulateRecordsToDeleteStopper = make(chan any)

// Create creates ShortURL.
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

// CreateMultiple creates multiple ShortURLs.
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

// GetByID returns ShortURL by its id.
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

// GetByUserID returns ShortURLs by user id.
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

// DeleteRecords deletes ShortURLs by ids.
func (repo *DatabaseRepository) DeleteRecords(ctx context.Context, userID uuid.UUID, ids []string) error {
	itemToDelete := &entities.ItemToDelete{
		UserID:   userID,
		ItemsIDs: ids,
	}
	toDeleteWaitGroup.Add(1)
	go func() {
		defer toDeleteWaitGroup.Done()
		repo.ToDelete <- itemToDelete
	}()
	return nil
}

// AccumulateRecordsToDelete accumulates ShortURLs to delete in background.
func (repo *DatabaseRepository) AccumulateRecordsToDelete() {
	ticker := time.NewTicker(time.Millisecond * 1500)
	defer ticker.Stop()

	localStorage := make(map[uuid.UUID][]string)

	go func() {
		defer repo.DeleteAccumulatorWaitGroup.Done()
		for {
			select {
			case <-ticker.C:
				repo.drainToDeleteStorage(localStorage)
			case <-accumulateRecordsToDeleteStopper:
				repo.drainToDeleteStorage(localStorage)
				log.Println("Finishing Accumulating coroutine")
				return
			}
		}
	}()

	for item := range repo.ToDelete {
		lockURLToDeleteStorage.Lock()
		urlsIDs, exist := localStorage[item.UserID]
		if exist {
			localStorage[item.UserID] = append(urlsIDs, item.ItemsIDs...)
		} else {
			localStorage[item.UserID] = item.ItemsIDs
		}
		lockURLToDeleteStorage.Unlock()
	}
}

func (repo *DatabaseRepository) drainToDeleteStorage(localStorage map[uuid.UUID][]string) {
	lockURLToDeleteStorage.Lock()
	for userID, ids := range localStorage {
		err := repo.DeleteRecordsForUser(context.Background(), userID, ids)
		if err != nil {
			log.Println("Error removing records")
		}
		delete(localStorage, userID)
	}
	lockURLToDeleteStorage.Unlock()
}

// CloseConnection closes database connection on request
func (repo *DatabaseRepository) CloseConnection() error {
	// waiting for all goroutines started in DeleteRecords
	toDeleteWaitGroup.Wait()
	close(repo.ToDelete)

	// waiting for ticker goroutine is being stopped by cancelled global context or by next line
	close(accumulateRecordsToDeleteStopper)

	// wait for accumulator goroutine stopped
	repo.DeleteAccumulatorWaitGroup.Wait()

	return repo.Storage.Close()
}

// DeleteRecordsForUser deletes all ShortURLs for user.
func (repo *DatabaseRepository) DeleteRecordsForUser(ctx context.Context, userID uuid.UUID, ids []string) error {
	query, args, _ := sqlx.In(
		"UPDATE short_urls SET is_active=false WHERE user_id = ? AND id IN (?)",
		userID.String(),
		ids,
	)
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	_, err := repo.Storage.ExecContext(
		ctx,
		query,
		args...,
	)
	return err
}

// GetOverallURLsAmount gets amount of urls.
func (repo *DatabaseRepository) GetOverallURLsAmount(ctx context.Context) (int, error) {
	var count int
	row := repo.Storage.QueryRowContext(
		ctx,
		"SELECT COUNT(original_url) as amount FROM short_urls",
	)
	err := row.Scan(&count)
	if err != nil {
		return count, err
	}
	return count, nil
}

// GetOverallUsersAmount gets amount of users.
func (repo *DatabaseRepository) GetOverallUsersAmount(ctx context.Context) (int, error) {
	var count int
	row := repo.Storage.QueryRowContext(
		ctx,
		"SELECT COUNT(DISTINCT user_id) as amount FROM short_urls",
	)
	err := row.Scan(&count)
	if err != nil {
		return count, err
	}
	return count, nil
}
