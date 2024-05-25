package algorithm

import "m1pes/internal/models"

func Algorithm(currentPrice float64, coin *models.Coin, user *models.User) {
	if currentPrice > coin.EntryPrice {
		coin.EntryPrice = currentPrice
		coin.Decrement = coin.EntryPrice * user.Percent * float64(len(coin.Buy)+1)
	}
	if coin.EntryPrice-coin.Decrement >= currentPrice {
		coin.Buy = append(coin.Buy, currentPrice) // покупать count * currentPrice
		if coin.Count == 0 {
			coin.Count += int64(user.Balance * 0.01 / currentPrice)
		} else {
			coin.Count = coin.Count / int64(len(coin.Buy)-1) * int64(len(coin.Buy))
		}
		coin.Decrement += coin.EntryPrice * user.Percent
	}
	var sum float64
	for i := 0; i < len(coin.Buy); i++ {
		sum += (coin.Buy)[i]
	}
	avg := sum / float64(len(coin.Buy))
	if avg+user.Percent*avg <= currentPrice {
		//sell
		clear(coin.Buy) // очистить в дб
		coin.Decrement = coin.EntryPrice * user.Percent
		coin.EntryPrice = 0
		coin.Count = 0
	}
}
