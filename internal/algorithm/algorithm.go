package algorithm

func Algorithm(currentPrice, bal float64, count *int64, buy []*float64, entryPrice, decrement, procent *float64) {
	//var startCopitalDay float64 // баланс клиента в 00:00 (в дб)
	//var bal float64          // потом уберем будет брать запросом с апи (нынешний баланс)
	//вводить в ручную клиентом (в дб)
	//var income float64     // map(int64)float64 (запихнем в дб)

	if currentPrice > *entryPrice {
		*entryPrice = currentPrice
		*decrement = *entryPrice * *procent * float64(len(buy)+1)

	}
	if *entryPrice-*decrement >= currentPrice {
		buy = append(buy, &currentPrice) // покупать count * currentPrice
		if *count == 0 {
			*count += int64(bal * 0.01 / currentPrice)
		} else {
			*count = *count / int64(len(buy)-1) * int64(len(buy))
		}
		*decrement += *entryPrice * *procent
	}
	var sum float64
	for i := 0; i < len(buy); i++ {
		sum += *buy[i]
	}
	avg := sum / float64(len(buy))
	if avg+*procent*avg <= currentPrice {
		//sell
		clear(buy) // очистить в дб
		*decrement = *entryPrice * *procent
		*entryPrice = 0
		*count = 0
	}
}
