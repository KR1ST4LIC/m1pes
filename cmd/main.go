package main

import (
	"context"
	"log"

	"m1pes/internal/app"
)

const (
	BTC  = "BTCUSDT"
	TON  = "TONUSDT"
	MEME = "MEMEUSDT"
	ETH  = "ETHUSDT"
	LTC  = "LTCUSDT"
)

type Coin struct {
	Decrement  float64
	EntryPrice float64
	Count      int64
	Buy        []*float64
}

func main() {

	ctx := context.Background()

	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err = a.Start(ctx); err != nil {
		log.Fatal(err)
	}

	//var bal float64
	//var procent float64
	//fmt.Scan(&bal)
	//fmt.Scan(&procent)
	//procent = procent * 0.01
	//monetki := []string{
	//	"BTCUSDT", "TONUSDT", "MEMEUSDT", "ETHUSDT", "LTCUSDT",
	//}
	//coins := make(map[string]Coin)
	//for i := 0; i < len(monetki); i++ {
	//	Price := GetPrice(monetki[i])
	//	coins[monetki[i]] = Coin{
	//		Decrement:  Price * procent,
	//		EntryPrice: Price,
	//		Buy:        make([]*float64, 0),
	//	}
	//}
	//for {
	//	for i := 0; i < len(monetki); i++ {
	//		currentPrice := GetPrice(monetki[i])
	//		count := coins[monetki[i]].Count
	//		buy := coins[monetki[i]].Buy
	//		decrement := coins[monetki[i]].Decrement
	//		entryPrice := coins[monetki[i]].EntryPrice
	//		algorithm.Algorithm(currentPrice, bal, &count, buy, &entryPrice, &decrement, &procent)
	//		fmt.Println(monetki[i], "    ", entryPrice, "    ", count)
	//		coins[monetki[i]] = Coin{
	//			Decrement:  decrement,
	//			EntryPrice: entryPrice,
	//			Buy:        buy,
	//			Count:      count,
	//		}
	//	}
	//}
}
