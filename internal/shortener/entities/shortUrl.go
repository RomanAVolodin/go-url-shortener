// Package entities stores all types of entities and DTOs.
package entities

import (
	"github.com/google/uuid"
)

// ShortURL main DTO to store entity in database.
type ShortURL struct {
	ID            string    `json:"id"`
	Short         string    `json:"short_url"`
	Original      string    `json:"original_url"`
	CorrelationID string    `json:"correlation_id"`
	UserID        uuid.UUID `json:"user_id"`
	IsActive      bool      `json:"is_active"`
}

// ShortURLResponseDto response dto.
type ShortURLResponseDto struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

// ShortURLResponseWithCorrelationDto response dto with correlation.
type ShortURLResponseWithCorrelationDto struct {
	CorrelationID string `json:"correlation_id"`
	Short         string `json:"short_url"`
}

// ShortURLWithCorrelationCreateDto dto for POST request.
type ShortURLWithCorrelationCreateDto struct {
	CorrelationID string `json:"correlation_id"`
	Original      string `json:"original_url"`
}

// ToResponseDto converts ShortURL to ShortURLResponseDto
func (item *ShortURL) ToResponseDto() ShortURLResponseDto {
	return ShortURLResponseDto{
		Short:    item.Short,
		Original: item.Original,
	}
}

// ToResponseWithCorrelationDto converts ShortURL to ShortURLResponseWithCorrelationDto
func (item *ShortURL) ToResponseWithCorrelationDto() ShortURLResponseWithCorrelationDto {
	return ShortURLResponseWithCorrelationDto{
		CorrelationID: item.CorrelationID,
		Short:         item.Short,
	}
}

// ShortenerSimpleCreateDTO simple create dto.
type ShortenerSimpleCreateDTO struct {
	URL string `json:"url"`
}

// ShortenerSimpleResponseDTO simple response dto.
type ShortenerSimpleResponseDTO struct {
	Result string `json:"result"`
}

// ItemToDelete delete dto.
type ItemToDelete struct {
	UserID   uuid.UUID
	ItemsIDs []string
}
