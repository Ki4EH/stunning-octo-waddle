package models

type User struct {
	Coin        int64       `json:"coin"`
	Inventory   []UserItem  `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}
