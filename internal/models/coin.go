package models

type Coin struct {
	UserId     int64
	Name       string
	EntryPrice float64
	Decrement  float64
	Count      int64
	Buy        []float64
}
