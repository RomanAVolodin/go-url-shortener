package utils

import "github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"

func GenerateResultURL(id string) string {
	return config.Settings.BaseURL + "/" + id
}
