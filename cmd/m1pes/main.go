package main

import (
	"context"
	"log"

	"m1pes/internal/app"
)

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

//package main
//
//import (
//	"context"
//	"fmt"
//	"github.com/hirokisan/bybit/v2"
//	"log"
//)
//
//func main() {
//	wsClient := bybit.NewWebsocketClient().WithBaseURL("wss://stream.bybit.com/v5/public/spot").WithAuth("e6jg0dLQEagHAiBvk6", "2G9xgJsp1Cl5wocmhlYBt9DQn56oM0v3psd7")
//	svc, err := wsClient.V5().Public("spot")
//	if err != nil {
//		return
//	}
//
//	f, err := svc.SubscribeTrade(bybit.V5WebsocketPublicTradeParamKey{Symbol: "NOTUSDT"}, func(bybit.V5WebsocketPublicTradeResponse) error {
//		fmt.Println("herere")
//		return nil
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = svc.Ping()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = f()
//	if err != nil {
//		log.Fatal(err, "from f func")
//	}
//
//	errHandler := func(isWebsocketClosed bool, err error) {
//		log.Fatal(err)
//	}
//
//	err = svc.Start(context.Background(), errHandler)
//	if err != nil {
//		log.Fatal(err)
//	}
//}
