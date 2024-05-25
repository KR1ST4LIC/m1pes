package models

type Coin struct {
	UserId     int64
	Name       string    `db:"coin_name"`
	EntryPrice float64   `db:"entry_price"`
	Decrement  float64   `db:"decrement"`
	Count      int64     `db:"count"`
	Buy        []float64 `db:"buy"`
}
