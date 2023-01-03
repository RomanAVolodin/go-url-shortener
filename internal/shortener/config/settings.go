package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"os"
)

const (
	OnlyGetPostRequestAllowedError = "Only GET/POST requests allowed"
	RequestBodyEmptyError          = "Request body is empty"
	BadInputData                   = "Incorrect request body data"
	UnknownError                   = "Something bad's happened"
	NoURLFoundByID                 = "No url found by id"
)

type AppSettings struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

var Settings AppSettings

func init() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Использование консольной команды %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flagSet.StringVar(&Settings.ServerAddress, "a", "localhost:8080", "Server address with port")
	flagSet.StringVar(&Settings.BaseURL, "b", "http://localhost:8080", "Full featured base url")
	flagSet.StringVar(&Settings.FileStoragePath, "f", "", "File storage path")
	flagSet.Parse(os.Args[1:])

	err := env.Parse(&Settings)
	if err != nil {
		log.Fatal(err)
	}
}
