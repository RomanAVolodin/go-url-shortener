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

	userID, err := middlewares.GetUserIDFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL, err := h.saveToRepository(createDTO.URL, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseDTO := ShortenerResponseDTO{Result: shortURL.Short}
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

	userID, err := middlewares.GetUserIDFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL, err := h.saveToRepository(string(urlToEncode), userID)
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
	userID, err := middlewares.GetUserIDFromCookie(r)
	if err != nil {
		http.Error(w, config.NoUserIDProvided, http.StatusBadRequest)
		return
	}

	records := h.Repo.GetByUserID(userID)
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

func (h *ShortenerHandler) saveToRepository(urlToEncode string, userID uuid.UUID) (entities.ShortURL, error) {
	id := shortuuid.New()
	shortURL := entities.ShortURL{
		ID:       id,
		Short:    utils.GenerateResultURL(id),
		Original: urlToEncode,
		UserID:   userID,
	}
	return h.Repo.Create(shortURL)
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
