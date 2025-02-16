package models

import (
	"github.com/google/uuid"
)

type Transaction struct {
	ID       uuid.UUID `json:"-"`
	FromUser string    `json:"fromUser"`
	ToUser   string    `json:"toUser"`
	Amount   int64     `json:"amount"`
}

type ReceivedTransaction struct {
	FromUser string `json:"fromUser"`
	Amount   int64  `json:"amount"`
}

type SentTransaction struct {
	ToUser string `json:"toUser"`
	Amount int64  `json:"amount"`
}

type CoinHistory struct {
	Received []ReceivedTransaction `json:"received"`
	Sent     []SentTransaction     `json:"sent"`
}
