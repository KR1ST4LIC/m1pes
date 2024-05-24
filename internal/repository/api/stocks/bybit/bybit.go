package bybit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	priceURL = "https://api.bybit.com/v2/public/tickers?symbol="
)

type Repository struct {
	cli *http.Client
}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) GetPrice(coinTag string) (float64, error) {
	apiKey := "e6jg0dLQEagHAiBvk6"

	url := priceURL + coinTag

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0, err
	}

	req.Header.Set("X-BAPI-API-KEY", apiKey)
	resp, err := r.cli.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return 0, err
	}

	var data map[string]interface{}

	if err = json.Unmarshal(body, &data); err != nil {
		fmt.Println("Ошибка при распаковке JSON:", err)
		return 0, err
	}

	result := data["result"].([]interface{})
	if len(result) == 0 {
		fmt.Println("Результат пустой")
		return 0, nil
	}

	bidPrice := result[0].(map[string]interface{})["bid_price"].(string)
	price, _ := strconv.ParseFloat(bidPrice, 64)
	return price, nil
}

func (r *Repository) ExistCoin(coinTag string) (bool, error) {
	apiKey := "e6jg0dLQEagHAiBvk6"

	url := priceURL + coinTag

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false, err
	}

	req.Header.Set("X-BAPI-API-KEY", apiKey)
	resp, err := r.cli.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return false, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return false, err
	}

	var data map[string]interface{}

	if err = json.Unmarshal(body, &data); err != nil {
		fmt.Println("Ошибка при распаковке JSON:", err)
		return false, err
	}

	result := data["result"].([]interface{})
	if len(result) == 0 {
		fmt.Println("Результат пустой")
		return false, nil
	}

	return true, nil
}
