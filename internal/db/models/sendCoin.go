package models

type SendCoin struct {
	ToUser string `json:"toUser"`
	Amount int64  `json:"amount"`
}
