package utils

import (
	"context"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

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
		err = os.Remove("test.db")
		if err != nil {
			log.Println(err)
		}
	}()

	config.Settings.IsTestMode = true

	tests := []struct {
		name        string
		filepath    string
		databaseDSN string
		want        string
	}{
		{
			name:        "Check if in memory repo",
			filepath:    "",
			databaseDSN: "",
			want:        reflect.TypeOf(&repositories.InMemoryRepository{}).String(),
		},
		{
			name:        "Check if file repo",
			filepath:    "test.json",
			databaseDSN: "",
			want:        reflect.TypeOf(&repositories.FileRepository{}).String(),
		},
		{
			name:        "Check if database repo",
			databaseDSN: "test.db",
			want:        reflect.TypeOf(&repositories.DatabaseRepository{}).String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.databaseDSN != "" {
				config.Settings.DatabaseDSN = tt.databaseDSN
			} else {
				config.Settings.FileStoragePath = tt.filepath
			}
			got := SetRepository(context.Background())

			assert.Equal(t, reflect.TypeOf(got).String(), tt.want)
			got.DeleteRecords(context.Background(), uuid.New(), []string{uuid.New().String(), uuid.New().String()})
			time.Sleep(time.Second)
		})
	}
}
