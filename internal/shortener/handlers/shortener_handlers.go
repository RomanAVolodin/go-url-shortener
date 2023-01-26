package handlers

import (
	"context"
	"encoding/json"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/middlewares"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lithammer/shortuuid"
	"io"
	"net/http"
	"time"
)

func (h *ShortenerHandler) CreateJSONShortURLHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestBody, doneWithError := h.readBody(w, r)
	if doneWithError {
		return
	}
	var createDTO entities.ShortenerSimpleCreateDTO
	if err := json.Unmarshal(requestBody, &createDTO); err != nil || createDTO.URL == "" {
		http.Error(w, config.BadInputData, http.StatusUnprocessableEntity)
		return
	}

	userID, err := middlewares.GetUserIDFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL, err := h.saveToRepository(createDTO.URL, userID, r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseDTO := entities.ShortenerSimpleResponseDTO{Result: shortURL.Short}
	jsonResponse, err := json.Marshal(responseDTO)
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func (h *ShortenerHandler) CreateMultipleShortURLHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestBody, doneWithError := h.readBody(w, r)
	if doneWithError {
		return
	}
	var incomingDTOs []entities.ShortURLResponseWithCorrelationCreateDto
	if err := json.Unmarshal(requestBody, &incomingDTOs); err != nil {
		http.Error(w, config.BadInputData, http.StatusUnprocessableEntity)
		return
	}
	userID, err := middlewares.GetUserIDFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items, err := h.saveMultipleToRepository(incomingDTOs, userID, r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var result = make([]entities.ShortURLResponseWithCorrelationDto, 0, len(items))
	for _, url := range items {
		result = append(result, url.ToResponseWithCorrelationDto())
	}
	jsonResponse, err := json.Marshal(result)
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

	userID, err := middlewares.GetUserIDFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL, err := h.saveToRepository(string(urlToEncode), userID, r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(shortURL.Short))
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
	if urlItem, exist := h.Repo.GetByID(r.Context(), urlID); exist {
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
	userID, err := middlewares.GetUserIDFromCookie(r)
	if err != nil {
		http.Error(w, config.NoUserIDProvided, http.StatusBadRequest)
		return
	}

	records := h.Repo.GetByUserID(r.Context(), userID)
	if len(records) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	responseDTOs := make([]entities.ShortURLResponseDto, 0, 8)
	for _, shortURL := range records {
		responseDTOs = append(responseDTOs, shortURL.ToResponseDto())
	}

	jsonRecords, _ := json.Marshal(responseDTOs)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonRecords)
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusInternalServerError)
		return
	}
}

func (h *ShortenerHandler) PingDatabase(w http.ResponseWriter, r *http.Request) {
	if repo, ok := h.Repo.(*repositories.DatabaseRepository); ok {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		if err := repo.Storage.PingContext(ctx); err != nil {
			http.Error(w, config.NoConnectionToDatabase, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, config.NoConnectionToDatabase, http.StatusInternalServerError)
}

func (h *ShortenerHandler) saveToRepository(
	urlToEncode string,
	userID uuid.UUID,
	ctx context.Context,
) (entities.ShortURL, error) {
	id := shortuuid.New()
	shortURL := entities.ShortURL{
		ID:       id,
		Short:    utils.GenerateResultURL(id),
		Original: urlToEncode,
		UserID:   userID,
	}
	return h.Repo.Create(ctx, shortURL)
}

func (h *ShortenerHandler) saveMultipleToRepository(
	items []entities.ShortURLResponseWithCorrelationCreateDto,
	userID uuid.UUID,
	ctx context.Context,
) ([]entities.ShortURL, error) {
	urls := make([]entities.ShortURL, 0, len(items))
	for _, item := range items {
		id := shortuuid.New()
		shortURL := entities.ShortURL{
			ID:            id,
			Short:         utils.GenerateResultURL(id),
			Original:      item.Original,
			CorrelationID: item.CorrelationID,
			UserID:        userID,
		}
		urls = append(urls, shortURL)
	}
	return h.Repo.CreateMultiple(ctx, urls)
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
