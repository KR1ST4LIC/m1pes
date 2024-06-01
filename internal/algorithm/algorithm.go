package algorithm

import (
	"fmt"
	"strconv"

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
			coin.Count, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", coin.Count+user.Balance*0.015/currentPrice), 64)
		} else {
			coin.Count, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", coin.Count/float64(len(coin.Buy)-1)*float64(len(coin.Buy))), 64)
		}
		coin.Decrement += coin.EntryPrice * user.Percent
		return BuyAction
	}
	var sum float64
	for i := 0; i < len(coin.Buy); i++ {
		sum += coin.Buy[i]
	}
	avg := sum / float64(len(coin.Buy))
	if avg+user.Percent*avg <= currentPrice {
		//sell
		coin.Decrement = coin.EntryPrice * user.Percent

		coin.Income = (currentPrice - avg) * coin.Count
		return SellAction
	}
	return WaitAction
}
