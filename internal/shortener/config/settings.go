package config

import (
	"github.com/caarlos0/env/v6"
	"log"
)

const (
	OnlyGetPostRequestAllowedError = "Only GET/POST requests allowed"
	RequestBodyEmptyError          = "Request body is empty"
	BadInputData                   = "Incorrect request body data"
	UnknownError                   = "Something bad's happened"
	NoURLFoundByID                 = "No url found by id"
)

type AppSettings struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

var Settings AppSettings

func init() {
	err := env.Parse(&Settings)
	if err != nil {
		log.Fatal(err)
	}
}
