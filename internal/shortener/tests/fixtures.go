package tests

import (
	"encoding/json"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/utils"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid"
)

var ShortUrlIDFixture = shortuuid.New()
var UserIdFixture = uuid.New()

var ShortUrlFixture = entities.ShortUrl{
	Id:       ShortUrlIDFixture,
	Short:    utils.GenerateResultURL(ShortUrlIDFixture),
	Original: "https://ya.ru",
	UserId:   UserIdFixture,
}

var ShortUrlFixtureSecond = entities.ShortUrl{
	Id:       ShortUrlIDFixture,
	Short:    utils.GenerateResultURL(ShortUrlIDFixture),
	Original: "https://mail.ru",
	UserId:   UserIdFixture,
}

var JsonStorageWithOneElement, _ = json.Marshal([]entities.ShortUrlResponseDto{ShortUrlFixture.ToResponseDto()})
