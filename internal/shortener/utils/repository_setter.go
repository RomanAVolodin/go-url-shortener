// Package utils includes shared utils.
package utils

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// SetRepository is the main method to set type of database to use in application.
func SetRepository() repositories.IRepository {
	if config.Settings.DatabaseDSN != "" {
		db, err := sql.Open("pgx", config.Settings.DatabaseDSN)
		if err != nil {
			log.Fatal(config.NoConnectionToDatabase)
		}
		// defer db.Close()
		// note, we haven't deffered db.Close() at the init function since the connection will close after init.
		// you could close it at main or ommit it

		db.SetMaxOpenConns(20)
		db.SetMaxIdleConns(20)
		db.SetConnMaxIdleTime(time.Second * 30)
		db.SetConnMaxLifetime(time.Minute * 2)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err = db.PingContext(ctx); err != nil {
			log.Fatal(config.NoConnectionToDatabase)
		}
		_, err = db.ExecContext(
			ctx,
			`CREATE TABLE IF NOT EXISTS short_urls (
				id varchar(45) NOT NULL PRIMARY KEY, 
				short_url varchar(150) NOT NULL, 
				original_url varchar(255) NOT NULL UNIQUE, 
				correlation_id varchar(255), 
				is_active boolean default true, 
				user_id uuid NOT NULL
            )`,
		)
		if err != nil {
			log.Fatal(config.NoConnectionToDatabase)
		}
		log.Println("Postgres storage`s been  chosen")
		repo := &repositories.DatabaseRepository{
			Storage:  db,
			ToDelete: make(chan *entities.ItemToDelete, 16),
		}
		go repo.AccumulateRecordsToDelete()
		return repo
	}
	if config.Settings.FileStoragePath != "" {
		repo := repositories.FileRepository{
			Storage:  make(map[string]entities.ShortURL),
			FilePath: config.Settings.FileStoragePath,
		}

		if err := repo.Restore(); err == nil {
			log.Println("File storage`s been  chosen")
			return &repo
		}
		log.Println("Error while choosing file storage")
	}

	repo := repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)}

	log.Println("In memory storage`s been chosen")
	return &repo
}
