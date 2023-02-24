package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/middlewares"

	"github.com/google/uuid"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
)

func ExampleShortener_CreateJSONShortURLHandler() {
	w := httptest.NewRecorder()
	handler := NewShortener(&repositories.InMemoryRepository{
		Storage: map[string]entities.ShortURL{},
	})
	r, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(`{"url": "https://mail.ru"}`),
	)
	h := http.HandlerFunc(handler.CreateJSONShortURLHandler)
	h.ServeHTTP(w, r)
}

func ExampleShortener_CreateMultipleShortURLHandler() {
	w := httptest.NewRecorder()
	handler := NewShortener(&repositories.InMemoryRepository{
		Storage: map[string]entities.ShortURL{},
	})
	r, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/shorten/batch",
		strings.NewReader(`[
					{
						"correlation_id": "mail.ru",
						"original_url": "https://mail.ru"
					},
					{
						"correlation_id": "yandex.ru",
						"original_url": "https://yandex.ru"
					}
				]`),
	)
	h := http.HandlerFunc(handler.CreateMultipleShortURLHandler)
	h.ServeHTTP(w, r)
}

func ExampleShortener_CreateShortURLHandler() {
	w := httptest.NewRecorder()
	handler := NewShortener(&repositories.InMemoryRepository{
		Storage: map[string]entities.ShortURL{},
	})
	r, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/",
		strings.NewReader("https://mail.ru"),
	)
	h := http.HandlerFunc(handler.CreateShortURLHandler)
	h.ServeHTTP(w, r)
}

func ExampleShortener_RetrieveShortURLHandler() {
	w := httptest.NewRecorder()
	handler := NewShortener(&repositories.InMemoryRepository{
		Storage: map[string]entities.ShortURL{"some_id": {
			ID:            "some_id",
			Short:         "some_id",
			Original:      "https://mail.ru",
			CorrelationID: "",
			UserID:        uuid.UUID{},
			IsActive:      true,
		}},
	})
	r, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/some_id",
		nil,
	)
	h := http.HandlerFunc(handler.RetrieveShortURLHandler)
	h.ServeHTTP(w, r)
}

func ExampleShortener_GetUsersRecordsHandler() {
	userID, _ := uuid.NewUUID()
	w := httptest.NewRecorder()
	handler := NewShortener(&repositories.InMemoryRepository{
		Storage: map[string]entities.ShortURL{"some_id": {
			ID:            "some_id",
			Short:         "some_id",
			Original:      "https://mail.ru",
			CorrelationID: "",
			UserID:        userID,
			IsActive:      true,
		}},
	})
	r, _ := http.NewRequestWithContext(
		context.WithValue(context.Background(), middlewares.UserIDKey, userID.String()),
		http.MethodGet,
		"/api/user/urls",
		nil,
	)
	h := http.HandlerFunc(handler.GetUsersRecordsHandler)
	h.ServeHTTP(w, r)
}

func ExampleShortener_PingDatabase() {
	w := httptest.NewRecorder()
	handler := NewShortener(&repositories.DatabaseRepository{})
	r, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/ping",
		nil,
	)
	h := http.HandlerFunc(handler.PingDatabase)
	h.ServeHTTP(w, r)
}

func ExampleShortener_DeleteRecordsHandler() {
	userID, _ := uuid.NewUUID()
	w := httptest.NewRecorder()
	handler := NewShortener(&repositories.InMemoryRepository{})
	r, _ := http.NewRequestWithContext(
		context.WithValue(context.Background(), middlewares.UserIDKey, userID.String()),
		http.MethodDelete,
		"/api/user/urls",
		strings.NewReader(`["shdfkjhaouyoiuwy", "26oigoajgdkhjgai"]`),
	)
	h := http.HandlerFunc(handler.DeleteRecordsHandler)
	h.ServeHTTP(w, r)
}
