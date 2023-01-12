package handlers

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestFileStorageShortURLHandler(t *testing.T) {
	defer func() {
		err := os.RemoveAll("test.json")
		if err != nil {
			log.Fatal(err)
		}
		err = os.RemoveAll("back_test.json")
		if err != nil {
			log.Fatal(err)
		}
	}()

	type wanted struct {
		code              int
		exactResponse     string
		responseStartWith string
		locationHeader    string
	}
	tests := []struct {
		name         string
		requestURL   string
		requestType  string
		requestBody  string
		repo         *repositories.FileRepository
		backRepo     *repositories.FileRepository
		wantedResult wanted
	}{
		{
			name:        "URL link should be generated",
			requestType: http.MethodPost,
			requestBody: "https://ya.ru",
			repo:        &repositories.FileRepository{Storage: make(map[string]string), FilePath: "test.json"},
			backRepo:    &repositories.FileRepository{Storage: make(map[string]string), FilePath: "back_test.json"},
			wantedResult: wanted{
				code:              http.StatusCreated,
				responseStartWith: config.Settings.BaseURL,
			},
		},
		{
			name:        "URL link should be returned",
			requestURL:  "/qwerty",
			requestType: http.MethodGet,
			requestBody: "https://ya.ru",
			repo:        &repositories.FileRepository{Storage: map[string]string{"qwerty": "https://ya.ru"}, FilePath: "test.json"},
			backRepo:    &repositories.FileRepository{Storage: make(map[string]string), FilePath: "back_test.json"},
			wantedResult: wanted{
				code:           http.StatusTemporaryRedirect,
				locationHeader: "https://ya.ru",
			},
		},
		{
			name:        "URL link should not be found with wrong id",
			requestURL:  "/randomid",
			requestType: http.MethodGet,
			requestBody: "https://ya.ru",
			repo:        &repositories.FileRepository{Storage: map[string]string{"qwerty": "https://ya.ru"}, FilePath: "test.json"},
			backRepo:    &repositories.FileRepository{Storage: make(map[string]string), FilePath: "back_test.json"},
			wantedResult: wanted{
				code:          http.StatusNotFound,
				exactResponse: config.NoURLFoundByID,
			},
		},
		{
			name:        "JSON URL link should be generated",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten",
			requestBody: "{\"url\": \"https://mail.ru\"}",
			repo:        &repositories.FileRepository{Storage: make(map[string]string), FilePath: "test.json"},
			backRepo:    &repositories.FileRepository{Storage: make(map[string]string), FilePath: "back_test.json"},
			wantedResult: wanted{
				code:              http.StatusCreated,
				responseStartWith: "{\"result\":\"http://",
			},
		},
		{
			name:        "JSON should return error with empty body",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten",
			repo:        &repositories.FileRepository{Storage: make(map[string]string), FilePath: "test.json"},
			backRepo:    &repositories.FileRepository{Storage: make(map[string]string), FilePath: "back_test.json"},
			wantedResult: wanted{
				code:          http.StatusBadRequest,
				exactResponse: config.RequestBodyEmptyError,
			},
		},
		{
			name:        "JSON should return error with wrong body body",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten",
			requestBody: "{\"wrongfield\": \"https://mail.ru\"}",
			repo:        &repositories.FileRepository{Storage: make(map[string]string), FilePath: "test.json"},
			backRepo:    &repositories.FileRepository{Storage: make(map[string]string), FilePath: "back_test.json"},
			wantedResult: wanted{
				code:          http.StatusUnprocessableEntity,
				exactResponse: config.BadInputData,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.repo.Restore()
			_ = tt.backRepo.Restore()

			url := "/"
			if tt.requestURL != "" {
				url = tt.requestURL
			}
			var request *http.Request
			if tt.requestBody != "" {
				request = httptest.NewRequest(tt.requestType, url, strings.NewReader(tt.requestBody))
			} else {
				request = httptest.NewRequest(tt.requestType, url, nil)
			}
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			h := NewShortenerHandler(tt.repo, tt.backRepo)
			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			assert.Equal(t, tt.wantedResult.code, res.StatusCode)
			assert.Nil(t, err)
			if tt.wantedResult.exactResponse != "" {
				assert.Equal(t, tt.wantedResult.exactResponse, strings.Trim(string(resBody), "\n"))
			}
			if tt.wantedResult.responseStartWith != "" {
				assert.True(
					t,
					strings.HasPrefix(strings.Trim(string(resBody), "\n"), tt.wantedResult.responseStartWith),
				)
			}

			assert.Equal(t, tt.wantedResult.locationHeader, res.Header.Get("Location"))
		})
	}
}
