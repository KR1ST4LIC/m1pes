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

	//createReq := models.CreateOrderRequest{
	//	Category:    "spot",
	//	Side:        "Buy",
	//	Symbol:      "MEMEUSDT",
	//	OrderType:   "Limit",
	//	Qty:         "1.0",
	//	Price:       "0.029",
	//	TimeInForce: "GTC",
	//}
	//jsonData, err := json.Marshal(createReq)
	//if err != nil {
	//	slog.ErrorContext(context.Background(), "Error marshaling create order", err)
	//}
	//
	//method := "POST"
	//apiSecret := "4FtzrEEpz8UYRxDyg3vzcLw0SR48KpOdO5A5"
	//apiKey := "ZD2sewIvQg6deMclTN"
	//params := string(jsonData)
	//URL := "https://api.bybit.com"
	//endPoint := "/v5/order/create"
	//
	//timestamp := time.Now().UnixMilli()
	//hmac256 := hmac.New(sha256.New, []byte(apiSecret))
	//hmac256.Write([]byte(strconv.FormatInt(timestamp, 10) + apiKey + "5000" + params))
	//signature := hex.EncodeToString(hmac256.Sum(nil))
	//
	//request, err := http.NewRequest(method, URL+endPoint+"?"+params, nil)
	//if err != nil {
	//	fmt.Println(err)
	//	//return nil, errors.Wrap(err, "failed create new request")
	//}
	//
	//if method == "POST" {
	//	request, err = http.NewRequest(method, URL+endPoint, bytes.NewBuffer([]byte(params)))
	//	if err != nil {
	//		fmt.Println(err)
	//		//return nil, errors.Wrap(err, "failed create new request")
	//	}
	//}
	//
	//request.Header.Set("Content-Type", "application/json")
	//request.Header.Set("X-BAPI-API-KEY", apiKey)
	//request.Header.Set("X-BAPI-SIGN", signature)
	//request.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timestamp, 10))
	//request.Header.Set("X-BAPI-SIGN-TYPE", "2")
	//request.Header.Set("X-BAPI-RECV-WINDOW", "5000")
	//
	//cli := &http.Client{}
	//
	//resp, err := cli.Do(request)
	//if err != nil {
	//	fmt.Println(err)
	//	//return nil, errors.Wrap(err, "failed do request")
	//}
	//defer resp.Body.Close()
	//
	//data, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	fmt.Println(err)
	//	//return nil, errors.Wrap(err, "failed read body")
	//}
	//
	//fmt.Println(string(data))
	//
	////return data, err
}
