package handlers

import (
	"encoding/json"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

type ShortenerCreateDTO struct {
	Url string `json:"url"`
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
	if err := json.Unmarshal(requestBody, &createDTO); err != nil || createDTO.Url == "" {
		http.Error(w, config.BadInputData, http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusCreated)

	var id string
	if link, exist := h.BackRepo.GetByID(createDTO.Url); exist {
		id = link
	} else {
		id = h.saveToRepositories([]byte(createDTO.Url))
	}

	responseDTO := ShortenerResponseDTO{Result: config.BaseURL + id}
	jsonResponse, err := json.Marshal(responseDTO)
	if err != nil {
		http.Error(w, config.UnknownError, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
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

	if link, exist := h.BackRepo.GetByID(string(urlToEncode)); exist {
		w.Write([]byte(config.BaseURL + link))
		return
	}

	id := h.saveToRepositories(urlToEncode)

	_, err := w.Write([]byte(config.BaseURL + id))
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

func (h *ShortenerHandler) saveToRepositories(urlToEncode []byte) string {
	id := h.Repo.Create(string(urlToEncode))
	h.BackRepo.CreateBackwardRecord(string(urlToEncode), id)
	return id
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
