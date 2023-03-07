// Package tests contains fixtures.
package tests

import (
	"encoding/json"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/utils"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid"
)

// ShortURLIDFixture short url fixture.
var ShortURLIDFixture = shortuuid.New()

// UserIDFixture user id fixture.
var UserIDFixture = uuid.New()

// ShortURLFixture ShortURL fixture.
var ShortURLFixture = entities.ShortURL{
	ID:            ShortURLIDFixture,
	Short:         utils.GenerateResultURL(ShortURLIDFixture),
	Original:      "https://ya.ru",
	UserID:        UserIDFixture,
	CorrelationID: "correlation_id",
	IsActive:      true,
}

// ShortURLFixtureInactive inactive ShortURL fixture
var ShortURLFixtureInactive = entities.ShortURL{
	ID:            ShortURLIDFixture,
	Short:         utils.GenerateResultURL(ShortURLIDFixture),
	Original:      "https://ya.ru",
	UserID:        UserIDFixture,
	CorrelationID: "correlation_id",
	IsActive:      false,
}

// JSONStorageWithOneElement fixture storage.
var JSONStorageWithOneElement, _ = json.Marshal([]entities.ShortURLResponseDto{ShortURLFixture.ToResponseDto()})
