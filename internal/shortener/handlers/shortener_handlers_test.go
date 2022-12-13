package handlers

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortURLHandler(t *testing.T) {
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
		repo         repositories.UrlsRepository
		backRepo     repositories.UrlsRepository
		wantedResult wanted
	}{
		{
			name:        "Delete request should fail",
			requestType: http.MethodDelete,
			wantedResult: wanted{
				code:          http.StatusBadRequest,
				exactResponse: config.OnlyGetPostRequestAllowedError,
			},
		},
		{
			name:        "Put request should fail",
			requestType: http.MethodPut,
			wantedResult: wanted{
				code:          http.StatusBadRequest,
				exactResponse: config.OnlyGetPostRequestAllowedError,
			},
		},
		{
			name:        "Patch request should fail",
			requestType: http.MethodPatch,
			wantedResult: wanted{
				code:          http.StatusBadRequest,
				exactResponse: config.OnlyGetPostRequestAllowedError,
			},
		},
		{
			name:        "Url link should not be generated with empty body",
			requestType: http.MethodPost,
			wantedResult: wanted{
				code:          http.StatusBadRequest,
				exactResponse: config.RequestBodyEmptyError,
			},
		},
		{
			name:        "Url link should be generated",
			requestType: http.MethodPost,
			requestBody: "https://ya.ru",
			repo:        make(repositories.UrlsRepository),
			backRepo:    make(repositories.UrlsRepository),
			wantedResult: wanted{
				code:              http.StatusCreated,
				responseStartWith: config.BaseURL,
			},
		},
		{
			name:        "Url link should be returned while creating if already generated",
			requestType: http.MethodPost,
			requestBody: "https://ya.ru",
			repo:        make(repositories.UrlsRepository),
			backRepo:    repositories.UrlsRepository{"https://ya.ru": "qwerty"},
			wantedResult: wanted{
				code:          http.StatusCreated,
				exactResponse: config.BaseURL + "qwerty",
			},
		},
		{
			name:        "Url link should not be returned without id in query",
			requestType: http.MethodGet,
			wantedResult: wanted{
				code:          http.StatusBadRequest,
				exactResponse: config.NoIDWasFoundInURL,
			},
		},
		{
			name:        "Url link should be returned",
			requestURL:  "/qwerty",
			requestType: http.MethodGet,
			requestBody: "https://ya.ru",
			repo:        repositories.UrlsRepository{"qwerty": "https://ya.ru"},
			backRepo:    make(repositories.UrlsRepository),
			wantedResult: wanted{
				code:           http.StatusTemporaryRedirect,
				locationHeader: "https://ya.ru",
			},
		},
		{
			name:        "Url link should not be found with wrong id",
			requestURL:  "/randomid",
			requestType: http.MethodGet,
			requestBody: "https://ya.ru",
			repo:        repositories.UrlsRepository{"qwerty": "https://ya.ru"},
			backRepo:    make(repositories.UrlsRepository),
			wantedResult: wanted{
				code:          http.StatusNotFound,
				exactResponse: config.NoURLFoundByID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			h := ShortenerHandler(tt.repo, tt.backRepo)
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
