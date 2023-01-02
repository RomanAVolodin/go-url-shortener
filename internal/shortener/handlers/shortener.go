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
	Repo     repo.Repository
	BackRepo repo.Repository
}

func NewShortenerHandler(repo repo.Repository, backRepo repo.Repository) *ShortenerHandler {
	h := &ShortenerHandler{
		Mux:      chi.NewMux(),
		Repo:     repo,
		BackRepo: backRepo,
	}
	h.Use(middleware.RequestID)
	h.Use(middleware.RealIP)
	h.Use(middleware.Logger)
	h.Use(middleware.Recoverer)
	h.Use(mw.GzipMiddleware)
	h.Use(mw.RequestUnzip)

	h.Get("/{id}", h.RetrieveShortURLHandler)
	h.Post("/", h.CreateShortURLHandler)
	h.Post("/api/shorten", h.CreateJSONShortURLHandler)
	h.MethodNotAllowed(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, config.OnlyGetPostRequestAllowedError, http.StatusMethodNotAllowed)
	})
	return h
}
