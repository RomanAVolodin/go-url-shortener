package handlers

import (
	"encoding/json"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/middlewares"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid"
	"io"
	"net/http"
)

type ShortenerCreateDTO struct {
	URL string `json:"url"`
}

type ShortenerResponseDTO struct {
	Result string `json:"result"`
}

func (h *ShortenerHandler) CreateJSONShortURLHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestBody, doneWithError := h.readBody(w, r)
	if doneWithError {
		return
	}
	var createDTO ShortenerCreateDTO
	if err := json.Unmarshal(requestBody, &createDTO); err != nil || createDTO.URL == "" {
		http.Error(w, config.BadInputData, http.StatusUnprocessableEntity)
		return
	}

	userId, err := middlewares.GetUserIDFromCookie(r)

	shortUrl, err := h.saveToRepository(createDTO.URL, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseDTO := ShortenerResponseDTO{Result: shortUrl.Short}
	jsonResponse, err := json.Marshal(responseDTO)
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func (h *ShortenerHandler) CreateShortURLHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	urlToEncode, doneWithError := h.readBody(w, r)
	if doneWithError {
		return
	}
	w.WriteHeader(http.StatusCreated)

	userId, err := middlewares.GetUserIDFromCookie(r)

	shortUrl, err := h.saveToRepository(string(urlToEncode), userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(shortUrl.Short))
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
	if urlItem, exist := h.Repo.GetByID(urlID); exist {
		w.Header().Set("Location", urlItem.Original)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	http.Error(w, config.NoURLFoundByID, http.StatusNotFound)
}

func (h *ShortenerHandler) GetUsersRecordsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userId, err := middlewares.GetUserIDFromCookie(r)
	if err != nil {
		http.Error(w, config.NoUserIdProvided, http.StatusBadRequest)
		return
	}

	records := h.Repo.GetByUserId(userId)
	if len(records) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	responseDTOs := make([]entities.ShortUrlResponseDto, 0, 8)
	for _, shortUrl := range records {
		responseDTOs = append(responseDTOs, shortUrl.ToResponseDto())
	}

	jsonRecords, _ := json.Marshal(responseDTOs)

	_, err = w.Write(jsonRecords)
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusInternalServerError)
		return
	}
}

func (h *ShortenerHandler) saveToRepository(urlToEncode string, userId uuid.UUID) (entities.ShortUrl, error) {
	id := shortuuid.New()
	shortUrl := entities.ShortUrl{
		Id:       id,
		Short:    utils.GenerateResultURL(id),
		Original: urlToEncode,
		UserId:   userId,
	}
	return h.Repo.Create(shortUrl)
}

func (h *ShortenerHandler) readBody(w http.ResponseWriter, r *http.Request) (body []byte, doneWithError bool) {
	urlToEncode, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusBadRequest)
		return nil, true
	}
	defer r.Body.Close()

	if string(urlToEncode) == "" {
		http.Error(w, config.RequestBodyEmptyError, http.StatusBadRequest)
		return nil, true
	}
	return urlToEncode, false
}
