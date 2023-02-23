package handlers

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/middlewares"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	tLoc "github.com/RomanAVolodin/go-url-shortener/internal/shortener/tests"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
		cookie       string
		repo         repositories.Repository
		wantedResult wanted
	}{
		{
			name:        "Delete request should fail",
			requestType: http.MethodDelete,
			wantedResult: wanted{
				code:          http.StatusMethodNotAllowed,
				exactResponse: config.OnlyGetPostRequestAllowedError,
			},
		},
		{
			name:        "Put request should fail",
			requestType: http.MethodPut,
			wantedResult: wanted{
				code:          http.StatusMethodNotAllowed,
				exactResponse: config.OnlyGetPostRequestAllowedError,
			},
		},
		{
			name:        "Patch request should fail",
			requestType: http.MethodPatch,
			wantedResult: wanted{
				code:          http.StatusMethodNotAllowed,
				exactResponse: config.OnlyGetPostRequestAllowedError,
			},
		},
		{
			name:        "URL link should not be generated with empty body",
			requestType: http.MethodPost,
			wantedResult: wanted{
				code:          http.StatusBadRequest,
				exactResponse: config.RequestBodyEmptyError,
			},
		},
		{
			name:        "URL link should be generated",
			requestType: http.MethodPost,
			requestBody: "https://ya.ru",
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
			wantedResult: wanted{
				code:              http.StatusCreated,
				responseStartWith: config.Settings.BaseURL,
			},
		},
		{
			name:        "URL link should not be returned without id in query",
			requestType: http.MethodGet,
			wantedResult: wanted{
				code:          http.StatusMethodNotAllowed,
				exactResponse: config.OnlyGetPostRequestAllowedError,
			},
		},
		{
			name:        "URL link should be returned",
			requestURL:  "/" + tLoc.ShortURLFixture.ID,
			requestType: http.MethodGet,
			requestBody: tLoc.ShortURLFixture.Original,
			repo: &repositories.InMemoryRepository{
				Storage: map[string]entities.ShortURL{tLoc.ShortURLFixture.ID: tLoc.ShortURLFixture},
			},
			wantedResult: wanted{
				code:           http.StatusTemporaryRedirect,
				locationHeader: tLoc.ShortURLFixture.Original,
			},
		},
		{
			name:        "URL link should not be found with wrong id",
			requestURL:  "/randomid",
			requestType: http.MethodGet,
			requestBody: tLoc.ShortURLFixture.Original,
			repo: &repositories.InMemoryRepository{
				Storage: map[string]entities.ShortURL{tLoc.ShortURLFixture.ID: tLoc.ShortURLFixture},
			},
			wantedResult: wanted{
				code:          http.StatusNotFound,
				exactResponse: config.NoURLFoundByID,
			},
		},
		{
			name:        "JSON handler URL link should not be generated with empty body",
			requestURL:  "/api/shorten",
			requestType: http.MethodPost,
			wantedResult: wanted{
				code:          http.StatusBadRequest,
				exactResponse: config.RequestBodyEmptyError,
			},
		},
		{
			name:        "JSON URL link should be generated",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten",
			requestBody: "{\"url\": \"https://mail.ru\"}",
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
			wantedResult: wanted{
				code:              http.StatusCreated,
				responseStartWith: "{\"result\":\"http://",
			},
		},
		{
			name:        "JSON should return error with empty body",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten",
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
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
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
			wantedResult: wanted{
				code:          http.StatusUnprocessableEntity,
				exactResponse: config.BadInputData,
			},
		},
		{
			name:        "Ping database should return error as database does not exist",
			requestType: http.MethodGet,
			requestURL:  "/ping",
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
			wantedResult: wanted{
				code: http.StatusInternalServerError,
			},
		},
		{
			name:        "Multiple JSON URL links should be generated",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten/batch",
			requestBody: "[{\"correlation_id\": \"mail\",\"original_url\": \"https://mail.ru\"}]",
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
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
			repo: &repositories.InMemoryRepository{
				Storage: map[string]entities.ShortURL{tLoc.ShortURLFixture.ID: tLoc.ShortURLFixture},
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
			repo: &repositories.InMemoryRepository{
				Storage: map[string]entities.ShortURL{tLoc.ShortURLFixture.ID: tLoc.ShortURLFixture},
			},
			cookie: middlewares.GenerateCookieStringForUserID(tLoc.UserIDFixture),
			wantedResult: wanted{
				code: http.StatusAccepted,
			},
		},
		{
			name:        "Deleted url should not be returned with response code 410",
			requestURL:  "/" + tLoc.ShortURLFixtureInactive.ID,
			requestType: http.MethodGet,
			requestBody: tLoc.ShortURLFixture.Original,
			repo: &repositories.InMemoryRepository{
				Storage: map[string]entities.ShortURL{tLoc.ShortURLFixtureInactive.ID: tLoc.ShortURLFixtureInactive},
			},
			wantedResult: wanted{
				code: http.StatusGone,
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

func TestRequestUnzip(t *testing.T) {
	repo := &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)}
	request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader("{\"url\": \"https://mail.ru\"}"))
	request.Header = http.Header{
		"Content-Type":    {"application/x-www-form-urlencoded; param=value"},
		"Accept-Encoding": {"gzip"},
	}
	w := httptest.NewRecorder()
	h := NewShortenerHandler(repo)
	h.ServeHTTP(w, request)
	res := w.Result()
	defer res.Body.Close()
	_, err := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Nil(t, err)
	assert.Equal(t, "gzip", res.Header.Get("Content-Encoding"))
}

func TestZippedContent(t *testing.T) {
	repo := &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)}

	zipped, _ := compress([]byte("{\"url\": \"https://mail.ru\"}"))
	request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(zipped))
	request.Header = http.Header{
		"Content-Type":     {"application/x-www-form-urlencoded; param=value"},
		"Content-Encoding": {"gzip"},
	}
	w := httptest.NewRecorder()
	h := NewShortenerHandler(repo)
	h.ServeHTTP(w, request)
	res := w.Result()
	defer res.Body.Close()
	_, err := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Nil(t, err)
}

func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	return b.Bytes(), nil
}

func TestAuthCookies(t *testing.T) {
	type wanted struct {
		cookieString string
		response     string
		code         int
	}
	tests := []struct {
		name         string
		requestURL   string
		requestType  string
		requestBody  string
		cookie       string
		repo         repositories.Repository
		wantedResult wanted
	}{
		{
			name:        "Request without auth cookie should obtain a new one in response",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten",
			requestBody: "{\"url\": \"https://mail.ru\"}",
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
			wantedResult: wanted{
				cookieString: "",
				code:         http.StatusCreated,
			},
		},
		{
			name:        "Request with auth cookie should return the same",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten",
			requestBody: "{\"url\": \"https://mail.ru\"}",
			cookie:      middlewares.GenerateCookieStringForUserID(tLoc.UserIDFixture),
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
			wantedResult: wanted{
				cookieString: middlewares.GenerateCookieStringForUserID(tLoc.UserIDFixture),
				code:         http.StatusCreated,
			},
		},
		{
			name:        "Request with incorrect auth cookie should return correct one",
			requestType: http.MethodPost,
			requestURL:  "/api/shorten",
			requestBody: "{\"url\": \"https://mail.ru\"}",
			cookie:      "wrong_cookie",
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
			wantedResult: wanted{
				code: http.StatusCreated,
			},
		},
		{
			name:        "User should receive empty answer",
			requestType: http.MethodGet,
			requestURL:  "/api/user/urls",
			cookie:      middlewares.GenerateCookieStringForUserID(tLoc.UserIDFixture),
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
			wantedResult: wanted{
				code: http.StatusNoContent,
			},
		},
		{
			name:        "User should receive his records",
			requestType: http.MethodGet,
			requestURL:  "/api/user/urls",
			cookie:      middlewares.GenerateCookieStringForUserID(tLoc.UserIDFixture),
			repo: &repositories.InMemoryRepository{
				Storage: map[string]entities.ShortURL{tLoc.ShortURLFixture.ID: tLoc.ShortURLFixture},
			},
			wantedResult: wanted{
				code:     http.StatusOK,
				response: string(tLoc.JSONStorageWithOneElement),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.requestType, tt.requestURL, strings.NewReader(tt.requestBody))
			request.Header = http.Header{
				"Content-Type": {"application/x-www-form-urlencoded; param=value"},
			}
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
			cookies := res.Cookies()
			var cookieReceived *http.Cookie
			for _, c := range cookies {
				if c.Name == middlewares.CookieName {
					cookieReceived = c
				}
			}
			defer res.Body.Close()
			resBody, _ := io.ReadAll(res.Body)

			assert.Equal(t, tt.wantedResult.code, res.StatusCode)
			assert.NotNil(t, cookieReceived)
			if tt.wantedResult.cookieString != "" {
				assert.Equal(t, tt.wantedResult.cookieString, cookieReceived.Value)
			}
			if tt.wantedResult.response != "" {
				assert.Equal(t, tt.wantedResult.response, strings.Trim(string(resBody), "\n"))
			}
		})
	}
}

func TestDatabaseRepository(t *testing.T) {
	type wanted struct {
		code               int
		responseBodyPrefix string
		header             string
	}
	tests := []struct {
		name       string
		urlString  string
		bodyString string
		method     string
		expect     func(sqlmock.Sqlmock)
		wanted     wanted
	}{
		{
			name:       "Create short URL success",
			urlString:  "/api/shorten",
			bodyString: "{\"url\": \"https://mail.ru\"}",
			method:     http.MethodPost,
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO short_urls").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wanted: wanted{code: http.StatusCreated, responseBodyPrefix: "{\"result\":\"http://"},
		},
		{
			name:       "Create short URL should return error with wrong params",
			urlString:  "/api/shorten",
			bodyString: "{\"url\": \"https://mail.ru\"}",
			method:     http.MethodPost,
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO short_urls").
					WillReturnError(errors.New("error while inserting"))
			},
			wanted: wanted{code: http.StatusInternalServerError},
		},
		{
			name:      "Get short URL by id success",
			urlString: "/some_id",
			method:    http.MethodGet,
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, short_url, original_url, user_id, correlation_id, is_active FROM short_urls").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "short_url", "original_url", "user_id", "correlation_id", "is_active"}).
							AddRow(
								tLoc.ShortURLFixture.ID,
								tLoc.ShortURLFixture.Short,
								tLoc.ShortURLFixture.Original,
								tLoc.ShortURLFixture.UserID.String(),
								tLoc.ShortURLFixture.CorrelationID,
								tLoc.ShortURLFixture.IsActive,
							),
					)
			},
			wanted: wanted{
				code:   http.StatusTemporaryRedirect,
				header: tLoc.ShortURLFixture.Original,
			},
		},
		{
			name:      "Get short URL by id should return error",
			urlString: "/some_id",
			method:    http.MethodGet,
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, short_url, original_url, user_id, correlation_id, is_active FROM short_urls").
					WillReturnError(sql.ErrNoRows)
			},
			wanted: wanted{
				code: http.StatusNotFound,
			},
		},
		{
			name:      "Get list by user id success",
			urlString: "/api/user/urls",
			method:    http.MethodGet,
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, short_url, original_url, user_id, correlation_id, is_active FROM short_urls").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "short_url", "original_url", "user_id", "correlation_id", "is_active"}).
							AddRow(
								tLoc.ShortURLFixture.ID,
								tLoc.ShortURLFixture.Short,
								tLoc.ShortURLFixture.Original,
								tLoc.ShortURLFixture.UserID.String(),
								tLoc.ShortURLFixture.CorrelationID,
								tLoc.ShortURLFixture.IsActive,
							),
					)
			},
			wanted: wanted{
				code: http.StatusOK,
			},
		},
		{
			name:      "Ping database should return OK as database exists",
			method:    http.MethodGet,
			urlString: "/ping",
			expect:    func(mock sqlmock.Sqlmock) {},
			wanted: wanted{
				code: http.StatusOK,
			},
		},
		{
			name:       "Create multiple short URL success",
			urlString:  "/api/shorten/batch",
			bodyString: "[{\"correlation_id\": \"mail\",\"original_url\": \"https://mail.ru\"}]",
			method:     http.MethodPost,
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO short_urls").WillBeClosed()
				mock.ExpectExec("INSERT INTO short_urls").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wanted: wanted{code: http.StatusCreated, responseBodyPrefix: "[{"},
		},
		{
			name:       "Create short URL that already exists should return it and status 409",
			urlString:  "/api/shorten",
			bodyString: "{\"url\": \"https://mail.ru\"}",
			method:     http.MethodPost,
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO short_urls").
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
				mock.ExpectQuery("SELECT id, short_url, original_url, user_id, correlation_id, is_active FROM short_urls").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "short_url", "original_url", "user_id", "correlation_id", "is_active"}).
							AddRow(
								tLoc.ShortURLFixture.ID,
								tLoc.ShortURLFixture.Short,
								tLoc.ShortURLFixture.Original,
								tLoc.ShortURLFixture.UserID.String(),
								tLoc.ShortURLFixture.CorrelationID,
								tLoc.ShortURLFixture.IsActive,
							),
					)
			},
			wanted: wanted{code: http.StatusConflict},
		},
		{
			name:       "Delete record success",
			urlString:  "/api/user/urls",
			bodyString: "[\"" + tLoc.ShortURLFixture.ID + "\"]",
			method:     http.MethodDelete,
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE short_urls SET is_active=false WHERE user_id").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wanted: wanted{code: http.StatusAccepted},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.expect(mock)

			var request *http.Request
			if tt.method != "" {
				request = httptest.NewRequest(tt.method, tt.urlString, strings.NewReader(tt.bodyString))
			} else {
				request = httptest.NewRequest(tt.method, tt.urlString, nil)
			}
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			repo := repositories.DatabaseRepository{Storage: db}

			w := httptest.NewRecorder()
			h := NewShortenerHandler(&repo)
			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			assert.Equal(t, tt.wanted.code, res.StatusCode)
			assert.Nil(t, err)

			if tt.wanted.responseBodyPrefix != "" {
				assert.True(
					t,
					strings.HasPrefix(strings.Trim(string(resBody), "\n"), tt.wanted.responseBodyPrefix),
				)
			}

			assert.Equal(t, tt.wanted.header, res.Header.Get("Location"))
		})
	}
}

func BenchmarkNewShortenerHandler(b *testing.B) {
	config.Settings.IsTestMode = true
	tests := []struct {
		name        string
		requestURL  string
		requestType string
		requestBody string
		repo        repositories.Repository
	}{
		{
			name:        "Get url by its ID",
			requestURL:  "/" + tLoc.ShortURLFixture.ID,
			requestType: http.MethodGet,
			requestBody: "",
		},
		{
			name:        "URL link generation",
			requestURL:  "/api/shorten",
			requestType: http.MethodPost,
			requestBody: "https://ya.ru",
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
		},
		{
			name:        "URL link generation",
			requestURL:  "/",
			requestType: http.MethodPost,
			requestBody: "https://ya.ru",
			repo:        &repositories.InMemoryRepository{Storage: make(map[string]entities.ShortURL)},
		},
	}
	w := httptest.NewRecorder()

	for _, tt := range tests {
		var request *http.Request
		if tt.requestBody != "" {
			request = httptest.NewRequest(
				tt.requestType,
				tt.requestURL,
				strings.NewReader(tt.requestBody),
			)
		} else {
			request = httptest.NewRequest(tt.requestType, tt.requestURL, nil)
		}
		h := NewShortenerHandler(&repositories.InMemoryRepository{
			Storage: map[string]entities.ShortURL{tLoc.ShortURLFixture.ID: tLoc.ShortURLFixture},
		})

		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				h.ServeHTTP(w, request)
				_ = w.Result()
			}
		})
	}
}
