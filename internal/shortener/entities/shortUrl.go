package entities

import (
	"github.com/google/uuid"
)

type ShortURL struct {
	ID       string    `json:"id"`
	Short    string    `json:"short_url"`
	Original string    `json:"original_url"`
	UserID   uuid.UUID `json:"user_id"`
}

type ShortURLResponseDto struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

func (item *ShortURL) ToResponseDto() ShortURLResponseDto {
	return ShortURLResponseDto{
		Short:    item.Short,
		Original: item.Original,
	}
}
