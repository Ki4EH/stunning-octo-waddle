package models

import "github.com/google/uuid"

type UserItem struct {
	UserID   uuid.UUID `json:"-"`
	Type     string    `json:"type"`
	Quantity int64     `json:"quantity"`
}
