// Package handlers is the main webserver.
package handlers

import (
	"net/http"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	mw "github.com/RomanAVolodin/go-url-shortener/internal/shortener/middlewares"
	repo "github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Shortener is struct based on Chi router with repository.
//
// https://github.com/go-chi/chi
type Shortener struct {
	*chi.Mux
	Repo repo.IRepository
}

// NewShortener creates new Shortener instance with all needed.
func NewShortener(repo repo.IRepository) *Shortener {
	h := &Shortener{
		Mux:  chi.NewMux(),
		Repo: repo,
	}
	h.Use(middleware.RequestID)
	h.Use(middleware.RealIP)
	if !config.Settings.IsTestMode {
		h.Use(middleware.Logger)
	}
	h.Use(middleware.Recoverer)
	h.Use(mw.GzipMiddleware)
	h.Use(mw.RequestUnzip)
	h.Use(mw.AuthCookie)

	h.Get("/{id}", h.RetrieveShortURLHandler)
	h.Post("/", h.CreateShortURLHandler)
	h.Post("/api/shorten", h.CreateJSONShortURLHandler)
	h.Post("/api/shorten/batch", h.CreateMultipleShortURLHandler)
	h.Get("/api/user/urls", h.GetUsersRecordsHandler)
	h.Delete("/api/user/urls", h.DeleteRecordsHandler)
	h.Get("/ping", h.PingDatabase)
	h.MethodNotAllowed(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, config.OnlyGetPostRequestAllowedError, http.StatusMethodNotAllowed)
	})

	if config.Settings.IsTestMode {
		h.Mount("/debug", middleware.Profiler())
	}
	return h
}
