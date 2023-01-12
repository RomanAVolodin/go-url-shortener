package handlers

import (
	"encoding/json"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"github.com/go-chi/chi/v5"
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

	var id string
	var err error
	if link, exist := h.BackRepo.GetByID(createDTO.URL); exist {
		id = link
	} else {
		id, err = h.saveToRepositories([]byte(createDTO.URL))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	responseDTO := ShortenerResponseDTO{Result: h.generateResultURL(r, id)}
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

	if link, exist := h.BackRepo.GetByID(string(urlToEncode)); exist {
		w.Write([]byte(h.generateResultURL(r, link)))
		return
	}

	id, err := h.saveToRepositories(urlToEncode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(h.generateResultURL(r, id)))
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

func (h *ShortenerHandler) saveToRepositories(urlToEncode []byte) (string, error) {
	id, err := h.Repo.CreateSave(string(urlToEncode))
	if err != nil {
		return "", err
	}
	_, err = h.BackRepo.Save(string(urlToEncode), id)
	return id, err
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

func (h *ShortenerHandler) generateResultURL(r *http.Request, id string) string {
	return config.Settings.BaseURL + "/" + id
}
