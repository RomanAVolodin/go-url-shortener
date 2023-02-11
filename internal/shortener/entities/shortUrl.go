package entities

import (
	"github.com/google/uuid"
)

type ShortURL struct {
	ID            string    `json:"id"`
	Short         string    `json:"short_url"`
	Original      string    `json:"original_url"`
	CorrelationID string    `json:"correlation_id"`
	UserID        uuid.UUID `json:"user_id"`
	IsActive      bool      `json:"is_active"`
}

type ShortURLResponseDto struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

type ShortURLResponseWithCorrelationDto struct {
	CorrelationID string `json:"correlation_id"`
	Short         string `json:"short_url"`
}

type ShortURLResponseWithCorrelationCreateDto struct {
	CorrelationID string `json:"correlation_id"`
	Original      string `json:"original_url"`
}

func (item *ShortURL) ToResponseDto() ShortURLResponseDto {
	return ShortURLResponseDto{
		Short:    item.Short,
		Original: item.Original,
	}
}

func (item *ShortURL) ToResponseWithCorrelationDto() ShortURLResponseWithCorrelationDto {
	return ShortURLResponseWithCorrelationDto{
		CorrelationID: item.CorrelationID,
		Short:         item.Short,
	}
}

type ShortenerSimpleCreateDTO struct {
	URL string `json:"url"`
}

type ShortenerSimpleResponseDTO struct {
	Result string `json:"result"`
}
