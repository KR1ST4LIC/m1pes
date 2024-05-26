package algorithm

import (
	"m1pes/internal/models"
)

const (
	ChangeAction = "change"
	SellAction   = "sell"
	BuyAction    = "buy"
	WaitAction   = "wait"
)

func Algorithm(currentPrice float64, coin *models.Coin, user *models.User) string {
	if currentPrice > coin.EntryPrice && coin.Count == 0 {
		coin.EntryPrice = currentPrice
		return ChangeAction
	}
	if coin.EntryPrice-coin.Decrement >= currentPrice {
		coin.Buy = append(coin.Buy, currentPrice) // покупать count * currentPrice
		if coin.Count == 0 {
			coin.Count += int64(user.Balance * 0.01 / currentPrice)
		} else {
			coin.Count = coin.Count / int64(len(coin.Buy)-1) * int64(len(coin.Buy))
		}
		coin.Decrement += coin.EntryPrice * user.Percent
		return BuyAction
	}
	var sum float64
	for i := 0; i < len(coin.Buy); i++ {
		sum += (coin.Buy)[i]
	}
	avg := sum / float64(len(coin.Buy))
	if avg+user.Percent*avg <= currentPrice {
		//sell
		coin.Decrement = coin.EntryPrice * user.Percent
		coin.EntryPrice = 0
		coin.Count = 0
		return SellAction
	}
	return WaitAction
}
