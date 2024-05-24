package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"m1pes/algorithm"
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
	var bal float64
	var procent float64
	fmt.Scan(&bal)
	fmt.Scan(&procent)
	procent = procent * 0.01
	monetki := []string{
		"BTCUSDT", "TONUSDT", "MEMEUSDT", "ETHUSDT", "LTCUSDT",
	}
	coins := make(map[string]Coin)
	for i := 0; i < len(monetki); i++ {
		Price := GetPrice(monetki[i])
		coins[monetki[i]] = Coin{
			Decrement:  Price * procent,
			EntryPrice: Price,
			Buy:        make([]*float64, 0),
		}
	}
	for {
		for i := 0; i < len(monetki); i++ {
			currentPrice := GetPrice(monetki[i])
			count := coins[monetki[i]].Count
			buy := coins[monetki[i]].Buy
			decrement := coins[monetki[i]].Decrement
			entryPrice := coins[monetki[i]].EntryPrice
			algorithm.Algorithm(currentPrice, bal, &count, buy, &entryPrice, &decrement, &procent)
			fmt.Println(monetki[i], "    ", entryPrice, "    ", count)
			coins[monetki[i]] = Coin{
				Decrement:  decrement,
				EntryPrice: entryPrice,
				Buy:        buy,
				Count:      count,
			}
		}
	}
}

func GetPrice(coinTag string) float64 {
	apiKey := "e6jg0dLQEagHAiBvk6"
	//secretKey := "YOUR_SECRET_KEY"

	// Создаем HTTP-клиент
	client := &http.Client{}

	// Формируем URL запроса к API Bybit для получения данных о котировках биткоина
	url := "https://api.bybit.com/v2/public/tickers?symbol=" + coinTag

	// Создаем HTTP-запрос

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0
	}

	// Добавляем заголовок для аутентификации
	req.Header.Set("X-BAPI-API-KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return 0
	}
	defer resp.Body.Close()

	// Читаем содержимое ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return 0
	}

	// Выводим содержимое ответа (JSON)
	var data map[string]interface{}

	// Распаковываем JSON
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Ошибка при распаковке JSON:", err)
		return 0
	}

	// Извлекаем значение bid_price
	result := data["result"].([]interface{})
	if len(result) == 0 {
		fmt.Println("Результат пустой")
		return 0
	}

	bidPrice := result[0].(map[string]interface{})["bid_price"].(string)
	price, _ := strconv.ParseFloat(bidPrice, 64)
	return price
}

//type response struct {
//	result []data `json:"result"`
//}
//
//type data struct {
//	bidPrice string `json:"bid_price"`
//}
