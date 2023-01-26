package tests

import (
	"encoding/json"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/utils"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid"
)

var ShortURLIDFixture = shortuuid.New()
var UserIDFixture = uuid.New()

var ShortURLFixture = entities.ShortURL{
	ID:            ShortURLIDFixture,
	Short:         utils.GenerateResultURL(ShortURLIDFixture),
	Original:      "https://ya.ru",
	UserID:        UserIDFixture,
	CorrelationID: "correlation_id",
}

var JSONStorageWithOneElement, _ = json.Marshal([]entities.ShortURLResponseDto{ShortURLFixture.ToResponseDto()})
