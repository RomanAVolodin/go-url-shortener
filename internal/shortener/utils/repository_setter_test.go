package utils

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
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
		err = os.Remove("test.json_back")
		if err != nil {
			log.Fatal(err)
		}
	}()

	tests := []struct {
		name     string
		filepath string
		want     repositories.Repository
		want1    repositories.Repository
	}{
		{
			name:     "Check if in memory repo",
			filepath: "",
			want:     &repositories.InMemoryRepository{Storage: make(map[string]string)},
			want1:    &repositories.InMemoryRepository{Storage: make(map[string]string)},
		},
		{
			name:     "Check if file repo",
			filepath: "test.json",
			want:     &repositories.FileRepository{Storage: make(map[string]string), FilePath: "test.json"},
			want1:    &repositories.FileRepository{Storage: make(map[string]string), FilePath: "test.json_back"},
		},
		{
			name:     "Check if file repo with incorrect filename should switch to in memory storage",
			filepath: "/some_unreal_name",
			want:     &repositories.InMemoryRepository{Storage: make(map[string]string)},
			want1:    &repositories.InMemoryRepository{Storage: make(map[string]string)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Settings.FileStoragePath = tt.filepath
			got, got1 := SetRepositories()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetRepositories() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("SetRepositories() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
