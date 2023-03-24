package utils

import (
	"context"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"github.com/stretchr/testify/assert"
)

func TestSetRepositories(t *testing.T) {
	defer func() {
		err := os.Remove("test.json")
		if err != nil {
			log.Println(err)
		}
	}()

	tests := []struct {
		name        string
		filepath    string
		databaseDSN string
		want        string
	}{
		{
			name:     "Check if in memory repo",
			filepath: "",
			want:     reflect.TypeOf(&repositories.InMemoryRepository{}).String(),
		},
		{
			name:     "Check if file repo",
			filepath: "test.json",
			want:     reflect.TypeOf(&repositories.FileRepository{}).String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Settings.FileStoragePath = tt.filepath
			config.Settings.DatabaseDSN = tt.databaseDSN
			got := SetRepository(context.Background())
			assert.Equal(t, reflect.TypeOf(got).String(), tt.want)
		})
	}
}
