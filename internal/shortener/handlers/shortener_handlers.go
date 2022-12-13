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
			CreateShortUrlHandler(repo, backRepo, w, r)
		case http.MethodGet:
			RetrieveShortUrlHandler(repo, w, r)
		default:
			http.Error(w, config.OnlyGetPostRequestAllowedError, http.StatusBadRequest)
		}
	}
}

func CreateShortUrlHandler(
	repo repositories.UrlsRepository,
	backRepo repositories.UrlsRepository,
	w http.ResponseWriter,
	r *http.Request,
) {
	defer r.Body.Close()
	urlToEncode, err := io.ReadAll(r.Body)
	if string(urlToEncode) == "" {
		http.Error(w, config.RequestBodyEmptyError, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)

	if link, exist := backRepo[string(urlToEncode)]; exist {
		_, err = w.Write([]byte(config.BaseUrl + link))
		return
	}

	shortUrl := shortuuid.New()
	repo[shortUrl] = string(urlToEncode)
	backRepo[string(urlToEncode)] = shortUrl

	_, err = w.Write([]byte(config.BaseUrl + shortUrl))
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusInternalServerError)
		return
	}
}

func RetrieveShortUrlHandler(
	repo repositories.UrlsRepository,
	w http.ResponseWriter,
	r *http.Request,
) {
	query := strings.Trim(r.URL.Path, "/")
	if query == "" {
		http.Error(w, config.NoIdWasFoundInUrl, http.StatusBadRequest)
		return
	}

	if link, exist := repo[query]; exist {
		w.Header().Set("Location", link)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	http.Error(w, config.NoUrlFoundById, http.StatusNotFound)
}
