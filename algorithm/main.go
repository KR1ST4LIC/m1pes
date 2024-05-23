package algorithm

import (
	"fmt"
)

func Algorithm() {
	//var startCopitalDay float64 // баланс клиента в 00:00 (в дб)
	var bal float64        // потом уберем будет брать запросом с апи (нынешний баланс)
	var priceMonet float64 // текущая цена монеты получаем запросом
	var procent float64    //вводить в ручную клиентом (в дб)
	//var income float64     // map(int64)float64 (запихнем в дб)
	var buy []float64 // запихиваем в логи
	var count int64
	var price float64
	var decrement float64

	fmt.Scan(&price)
	fmt.Scan(&bal)
	fmt.Scan(&procent)
	procent = procent * 0.01
	decrement = price * procent
	count = 0
	for {
		fmt.Scan(&priceMonet)
		if price-decrement >= priceMonet {
			buy = append(buy, priceMonet) // покупать count * priceMonet
			if count == 0 {
				count += int64(bal * 0.01 / priceMonet)
			} else {
				count = count / int64(len(buy)-1) * int64(len(buy))
			}
			decrement += price * procent
		}
		var sum float64
		for i := 0; i < len(buy); i++ {
			sum += buy[i]
		}
		avg := sum / float64(len(buy))
		if avg+procent*avg <= priceMonet {
			//sell
			clear(buy) // очистить в дб
			decrement = price * procent
			price = priceMonet
			count = 0
		}
	}

	//if time.Now().Hour() == 0 {
	//	startCopitalDay = bal
	//}

}
