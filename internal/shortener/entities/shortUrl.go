package entities

import "github.com/google/uuid"

type ShortUrl struct {
	Id       string    `json:"id"`
	Short    string    `json:"short_url"`
	Original string    `json:"original_url"`
	UserId   uuid.UUID `json:"user_id"`
}

type ShortUrlResponseDto struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

func (item *ShortUrl) ToResponseDto() ShortUrlResponseDto {
	return ShortUrlResponseDto{
		Short:    item.Short,
		Original: item.Original,
	}
}
