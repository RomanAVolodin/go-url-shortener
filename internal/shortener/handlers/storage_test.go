package handlers

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/middlewares"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	tLoc "github.com/RomanAVolodin/go-url-shortener/internal/shortener/tests"
	"github.com/stretchr/testify/assert"
)

func TestFileStorageShortURLHandler(t *testing.T) {
	defer func() {
		err := os.RemoveAll("test.json")
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
		cookie       string
		repo         *repositories.FileRepository
		wantedResult wanted
	}{
		{
			name:        "URL link should be generated",
			requestType: http.MethodPost,
			requestBody: "https://ya.ru",
			repo: &repositories.FileRepository{
				Storage:  make(map[string]entities.ShortURL),
				FilePath: "test.json",
			},
			wantedResult: wanted{
				code:              http.StatusCreated,
				responseStartWith: config.Settings.BaseURL,
			},
		},
		{
			name:        "URL link should be returned",
			requestURL:  "/" + tLoc.ShortURLFixture.ID,
			requestType: http.MethodGet,
			requestBody: tLoc.ShortURLFixture.Original,
			repo: &repositories.FileRepository{
				Storage:  map[string]entities.ShortURL{tLoc.ShortURLFixture.ID: tLoc.ShortURLFixture},
				FilePath: "test.json",
			},
			wantedResult: wanted{
				code:           http.StatusTemporaryRedirect,
				locationHeader: tLoc.ShortURLFixture.Original,
			},
		},
		{
			name:        "URLs list should be returned by user id",
			requestURL:  "/api/user/urls",
			requestType: http.MethodGet,
			repo: &repositories.FileRepository{
				Storage:  map[string]entities.ShortURL{tLoc.ShortURLFixture.ID: tLoc.ShortURLFixture},
				FilePath: "test.json",
			},
			wantedResult: wanted{
				code: http.StatusNoContent,
			},
		},
		{
			name:        "URL link should not be found with wrong id",
			requestURL:  "/randomid",
			requestType: http.MethodGet,
			requestBody: "https://ya.ru",
			repo: &repositories.FileRepository{
				Storage:  map[string]entities.ShortURL{tLoc.ShortURLFixture.ID: tLoc.ShortURLFixture},
				FilePath: "test.json",
			},
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
			repo: &repositories.FileRepository{
				Storage:  make(map[string]entities.ShortURL),
				FilePath: "test.json",
			},
			wantedResult: wanted{
				code:              http.StatusCreated,
				responseStartWith: "{\"result\":\"http://",
			},
		},
		{
			name:        "JSON should return error with empty body",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten",
			repo: &repositories.FileRepository{
				Storage:  make(map[string]entities.ShortURL),
				FilePath: "test.json",
			},
			wantedResult: wanted{
				code:          http.StatusBadRequest,
				exactResponse: config.RequestBodyEmptyError,
			},
		},
		{
			name:        "JSON should return error with wrong body",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten",
			requestBody: "{\"wrongfield\": \"https://mail.ru\"}",
			repo: &repositories.FileRepository{
				Storage:  make(map[string]entities.ShortURL),
				FilePath: "test.json",
			},
			wantedResult: wanted{
				code:          http.StatusUnprocessableEntity,
				exactResponse: config.BadInputData,
			},
		},
		{
			name:        "Multiple JSON URL link should be generated",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten/batch",
			requestBody: "[{\"correlation_id\": \"mail\",\"original_url\": \"https://mail.ru\"}]",
			repo: &repositories.FileRepository{
				Storage:  make(map[string]entities.ShortURL),
				FilePath: "test.json",
			},
			wantedResult: wanted{
				code:              http.StatusCreated,
				responseStartWith: "[{",
			},
		},
		{
			name:        "Delete users Urls should be success even for wrong user",
			requestURL:  "/api/user/urls",
			requestType: http.MethodDelete,
			requestBody: "[\"" + tLoc.ShortURLFixture.ID + "\"]",
			repo: &repositories.FileRepository{
				Storage:  make(map[string]entities.ShortURL),
				FilePath: "test.json",
			},
			wantedResult: wanted{
				code: http.StatusAccepted,
			},
		},
		{
			name:        "Delete users Urls should be success for owner",
			requestURL:  "/api/user/urls",
			requestType: http.MethodDelete,
			requestBody: "[\"" + tLoc.ShortURLFixture.ID + "\"]",
			repo: &repositories.FileRepository{
				Storage:  map[string]entities.ShortURL{tLoc.ShortURLFixture.ID: tLoc.ShortURLFixture},
				FilePath: "test.json",
			},
			cookie: middlewares.GenerateCookieStringForUserID(tLoc.UserIDFixture),
			wantedResult: wanted{
				code: http.StatusAccepted,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.repo.Restore()

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
			h := NewShortenerHandler(tt.repo)

			if tt.cookie != "" {
				cookie := &http.Cookie{
					Name:     middlewares.CookieName,
					Value:    tt.cookie,
					Expires:  time.Now().Add(24 * time.Hour),
					HttpOnly: true,
				}
				request.AddCookie(cookie)
			}

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
