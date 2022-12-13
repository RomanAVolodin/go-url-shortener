package handlers

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"github.com/lithammer/shortuuid"
	"io"
	"net/http"
	"strings"
)

func ShortenerHandler(repo repositories.UrlsRepository, backRepo repositories.UrlsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch method := r.Method; method {
		case http.MethodPost:
			CreateShortURLHandler(repo, backRepo, w, r)
		case http.MethodGet:
			RetrieveShortURLHandler(repo, w, r)
		default:
			http.Error(w, config.OnlyGetPostRequestAllowedError, http.StatusBadRequest)
		}
	}
}

func CreateShortURLHandler(
	repo repositories.UrlsRepository,
	backRepo repositories.UrlsRepository,
	w http.ResponseWriter,
	r *http.Request,
) {
	defer r.Body.Close()
	urlToEncode, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusBadRequest)
		return
	}
	if string(urlToEncode) == "" {
		http.Error(w, config.RequestBodyEmptyError, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)

	if link, exist := backRepo[string(urlToEncode)]; exist {
		w.Write([]byte(config.BaseURL + link))
		return
	}

	shortURL := shortuuid.New()
	repo[shortURL] = string(urlToEncode)
	backRepo[string(urlToEncode)] = shortURL

	_, err = w.Write([]byte(config.BaseURL + shortURL))
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusInternalServerError)
		return
	}
}

func RetrieveShortURLHandler(
	repo repositories.UrlsRepository,
	w http.ResponseWriter,
	r *http.Request,
) {
	query := strings.Trim(r.URL.Path, "/")
	if query == "" {
		http.Error(w, config.NoIDWasFoundInURL, http.StatusBadRequest)
		return
	}

	if link, exist := repo[query]; exist {
		w.Header().Set("Location", link)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	http.Error(w, config.NoURLFoundByID, http.StatusNotFound)
}
