package models

type Coin struct {
	UserId       int64
	Name         string
	EntryPrice   float64
	CurrentPrice float64
	Decrement    float64
	Count        float64
	Buy          []float64
}
