package models

type User struct {
	Coin        int64       `json:"coins"`
	Inventory   []UserItem  `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}
