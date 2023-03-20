// Package config holds all the configuration variables for application.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

// Constants for error responses.
const (
	OnlyGetPostRequestAllowedError = "Only GET/POST requests allowed"
	RequestBodyEmptyError          = "Request body is empty"
	BadInputData                   = "Incorrect request body data"
	UnknownError                   = "Something bad's happened"
	NoURLFoundByID                 = "No url found by id"
	NoUserIDProvided               = "No user ID has been provided"
	NoConnectionToDatabase         = "Error while connecting to database"
)

// AppSettings struct to handle application settings parsed from environment variables.
type AppSettings struct {
	ServerAddress   string `env:"SERVER_ADDRESS"    json:"server_address"`
	BaseURL         string `env:"BASE_URL"          json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	SecretAuthKey   string `env:"AUTH_SECRET_KEY"   default:"super_secret"`
	DatabaseDSN     string `env:"DATABASE_DSN"      json:"database_dsn"`
	IsTestMode      bool   `env:"IS_TEST"           default:"false"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"      json:"enable_https"`
	ConfigFile      string `env:"CONFIG"`
}

// Settings singleton with application configuration, initializes in `init()` method.
var Settings AppSettings

func parseConfigFile(settings *AppSettings) {
	var config AppSettings
	file, err := os.ReadFile(settings.ConfigFile)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return
	}

	if settings.ServerAddress == "" {
		settings.ServerAddress = config.ServerAddress
	}
	if settings.BaseURL == "" {
		settings.BaseURL = config.BaseURL
	}
	if settings.FileStoragePath == "" {
		settings.FileStoragePath = config.FileStoragePath
	}
	if settings.DatabaseDSN == "" {
		settings.DatabaseDSN = config.DatabaseDSN
	}
	if !settings.EnableHTTPS {
		settings.EnableHTTPS = config.EnableHTTPS
	}
}

func init() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Использование консольной команды %s:\n", os.Args[0])
		flagSet.PrintDefaults()
	}
	flagSet.StringVar(&Settings.ServerAddress, "a", "localhost:8080", "Server address with port")
	flagSet.StringVar(&Settings.BaseURL, "b", "http://localhost:8080", "Full featured base url")
	flagSet.StringVar(&Settings.FileStoragePath, "f", "", "File storage path")
	flagSet.StringVar(&Settings.DatabaseDSN, "d", "", "Database DSN url")
	flagSet.BoolVar(&Settings.EnableHTTPS, "s", false, "Enable HTTPs")
	flagSet.StringVar(&Settings.ConfigFile, "c", "", "Config file location")
	flagSet.StringVar(&Settings.ConfigFile, "config", "", "Config file location")
	flagSet.Parse(os.Args[1:])

	if Settings.ConfigFile != "" {
		parseConfigFile(&Settings)
	}

	err := env.Parse(&Settings)
	if err != nil {
		log.Fatal(err)
	}
}
