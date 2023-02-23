package handlers

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	mw "github.com/RomanAVolodin/go-url-shortener/internal/shortener/middlewares"
	repo "github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

type ShortenerHandler struct {
	*chi.Mux
	Repo repo.Repository
}

func NewShortenerHandler(repo repo.Repository) *ShortenerHandler {
	h := &ShortenerHandler{
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
