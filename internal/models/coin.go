package models

type Coin struct {
	UserId       int64
	Name         string
	EntryPrice   float64
	CurrentPrice float64
	Decrement    float64
	Count        float64
	Buy          []float64
	Income       float64
}

func NewCoin(userId int64, coinName string) Coin {
	coin := Coin{UserId: userId, Name: coinName}
	return coin
}

var CoinPrice = make(map[string]float64)

type List struct {
	Name  []string
	Buy   map[string][]float64
	Count []float64
}
