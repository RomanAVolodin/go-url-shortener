package utils

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestSetRepositories(t *testing.T) {
	defer func() {
		err := os.Remove("test.json")
		if err != nil {
			log.Fatal(err)
		}
	}()

	tests := []struct {
		name     string
		filepath string
		want     repositories.Repository
	}{
		{
			name:     "Check if in memory repo",
			filepath: "",
			want:     &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortUrl)},
		},
		{
			name:     "Check if file repo",
			filepath: "test.json",
			want:     &repositories.FileRepository{Storage: make(map[string]entities.ShortUrl), FilePath: "test.json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Settings.FileStoragePath = tt.filepath
			got := SetRepository()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetRepository() got = %v, want %v", got, tt.want)
			}
		})
	}
}
