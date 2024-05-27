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
}
