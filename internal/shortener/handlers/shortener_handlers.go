package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/middlewares"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/shortenerrors"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lithammer/shortuuid"
)

// CreateJSONShortURLHandler handles POST request with json DTO.
func (h *Shortener) CreateJSONShortURLHandler(
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

	userID := r.Context().Value(middlewares.UserIDKey).(uuid.UUID)

	shortURL, statusCode, err := h.saveToRepository(r.Context(), createDTO.URL, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	responseDTO := entities.ShortenerSimpleResponseDTO{Result: shortURL.Short}
	jsonResponse, err := json.Marshal(responseDTO)
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusBadRequest)
		return
	}
	w.Write(jsonResponse)
}

// CreateMultipleShortURLHandler handles POST request with multiple urls to process.
func (h *Shortener) CreateMultipleShortURLHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestBody, doneWithError := h.readBody(w, r)
	if doneWithError {
		return
	}
	var incomingDTOs []entities.ShortURLWithCorrelationCreateDto
	if err := json.Unmarshal(requestBody, &incomingDTOs); err != nil {
		http.Error(w, config.BadInputData, http.StatusUnprocessableEntity)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(uuid.UUID)

	items, err := h.saveMultipleToRepository(r.Context(), incomingDTOs, userID)
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

// CreateShortURLHandler handles simple POST requests with URL.
func (h *Shortener) CreateShortURLHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	urlToEncode, doneWithError := h.readBody(w, r)
	if doneWithError {
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(uuid.UUID)

	shortURL, statusCode, err := h.saveToRepository(r.Context(), string(urlToEncode), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)

	_, err = w.Write([]byte(shortURL.Short))
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusInternalServerError)
		return
	}
}

// RetrieveShortURLHandler returns short url by it`s id.
func (h *Shortener) RetrieveShortURLHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	urlID := chi.URLParam(r, "id")
	urlItem, exist, err := h.Repo.GetByID(r.Context(), urlID)

	if exist && err == nil {
		if !urlItem.IsActive {
			w.WriteHeader(http.StatusGone)
			return
		}
		w.Header().Set("Location", urlItem.Original)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if !exist || (err != nil && errors.Is(err, sql.ErrNoRows)) {
		http.Error(w, config.NoURLFoundByID, http.StatusNotFound)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// GetUsersRecordsHandler returns all records related to current user.
func (h *Shortener) GetUsersRecordsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID := r.Context().Value(middlewares.UserIDKey).(uuid.UUID)

	records, err := h.Repo.GetByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

// PingDatabase returns Database connection status
func (h *Shortener) PingDatabase(w http.ResponseWriter, r *http.Request) {
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

// DeleteRecordsHandler handles records deletion by it's IDs.
func (h *Shortener) DeleteRecordsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestBody, doneWithError := h.readBody(w, r)
	if doneWithError {
		return
	}
	var idsToDelete []string
	if err := json.Unmarshal(requestBody, &idsToDelete); err != nil || len(idsToDelete) == 0 {
		http.Error(w, config.BadInputData, http.StatusUnprocessableEntity)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(uuid.UUID)

	err := h.deleteFromRepository(r.Context(), idsToDelete, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetServiceStats returns the statistics.
func (h *Shortener) GetServiceStats(
	w http.ResponseWriter,
	r *http.Request,
) {
	_, network, err := net.ParseCIDR(config.Settings.TrustedSubnet)
	userIP := net.ParseIP(r.Header.Get("X-Real-IP"))
	if err == nil && !network.Contains(userIP) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	urlsAmount, err := h.Repo.GetOverallURLsAmount(r.Context())
	usersAmount, err := h.Repo.GetOverallUsersAmount(r.Context())
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusBadRequest)
		return
	}

	jsonResponse, err := json.Marshal(map[string]int{"urls": urlsAmount, "users": usersAmount})
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func (h *Shortener) deleteFromRepository(
	ctx context.Context,
	ids []string,
	userID uuid.UUID,
) error {
	return h.Repo.DeleteRecords(ctx, userID, ids)
}

func (h *Shortener) saveToRepository(
	ctx context.Context,
	urlToEncode string,
	userID uuid.UUID,
) (entities.ShortURL, int, error) {
	id := shortuuid.New()
	shortURL := entities.ShortURL{
		ID:       id,
		Short:    utils.GenerateResultURL(id),
		Original: urlToEncode,
		UserID:   userID,
		IsActive: true,
	}
	url, err := h.Repo.Create(ctx, shortURL)

	var statusCode int
	switch {
	case err != nil && errors.Is(err, shortenerrors.ErrItemAlreadyExists):
		statusCode = http.StatusConflict
	case err != nil:
		return url, statusCode, err
	default:
		statusCode = http.StatusCreated
	}
	return url, statusCode, nil
}

func (h *Shortener) saveMultipleToRepository(
	ctx context.Context,
	items []entities.ShortURLWithCorrelationCreateDto,
	userID uuid.UUID,
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
			IsActive:      true,
		}
		urls = append(urls, shortURL)
	}
	return h.Repo.CreateMultiple(ctx, urls)
}

func (h *Shortener) readBody(w http.ResponseWriter, r *http.Request) (body []byte, doneWithError bool) {
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
