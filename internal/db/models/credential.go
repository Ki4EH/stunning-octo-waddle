package models

import "github.com/google/uuid"

type Credential struct {
	ID       uuid.UUID
	Username string `json:"username"`
	Password string `json:"password"`
	Coin     int64  `json:"coin"`
}
