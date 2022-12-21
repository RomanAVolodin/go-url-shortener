package handlers

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	repo "github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
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

	h.Get("/{id}", h.RetrieveShortURLHandler)
	h.Post("/", h.CreateShortURLHandler)
	h.MethodNotAllowed(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, config.OnlyGetPostRequestAllowedError, http.StatusMethodNotAllowed)
	})
	return h
}

func (h *ShortenerHandler) CreateShortURLHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	urlToEncode, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if string(urlToEncode) == "" {
		http.Error(w, config.RequestBodyEmptyError, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)

	if link, exist := h.BackRepo.GetByID(string(urlToEncode)); exist {
		w.Write([]byte(config.BaseURL + link))
		return
	}

	// TODO: wrap into one transaction
	id := h.Repo.Create(string(urlToEncode))
	h.BackRepo.CreateBackwardRecord(string(urlToEncode), id)

	_, err = w.Write([]byte(config.BaseURL + id))
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusInternalServerError)
		return
	}
}

func (h *ShortenerHandler) RetrieveShortURLHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	urlID := chi.URLParam(r, "id")
	if link, exist := h.Repo.GetByID(urlID); exist {
		w.Header().Set("Location", link)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	http.Error(w, config.NoURLFoundByID, http.StatusNotFound)
}
