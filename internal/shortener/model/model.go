package model

import "time"

// URL represents the data for the URL.
//
//nolint:lll
type URL struct {
	ID             int64      `json:"id,omitempty"`
	ObjectID       string     `json:"objectID,omitempty" bson:"_id"`
	OriginalURL    string     `json:"originalURL" db:"original_url" bson:"original_url" validate:"required,url"`
	ShortURL       string     `json:"shortURL" db:"short_url" bson:"short_url"`
	CustomURL      *string    `json:"customURL" db:"custom_url" bson:"custom_url" validate:"omitempty,alphanum,min=3,max=20"`
	ExpirationDate *time.Time `json:"expirationDate" db:"expiration_date" bson:"expiration_date" validate:"omitempty"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at" bson:"created_at"`
}
