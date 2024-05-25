package bybit

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	priceURL = "https://api.bybit.com/v2/public/tickers?symbol="
	apiKey   = "e6jg0dLQEagHAiBvk6"
)

type Repository struct {
	cli *http.Client
}

func New() *Repository {
	return &Repository{
		cli: &http.Client{
			Timeout: time.Minute,
		},
	}
}

func (r *Repository) GetPrice(ctx context.Context, coinTag string) (float64, error) {
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
		fmt.Println("error in unmarshal:", err)
		return 0, err
	}

	result := data["result"].([]interface{})
	if len(result) == 0 {
		fmt.Println("empty result")
		return 0, nil
	}

	bidPrice := result[0].(map[string]interface{})["bid_price"].(string)
	price, _ := strconv.ParseFloat(bidPrice, 64)
	return price, nil
}

func (r *Repository) ExistCoin(ctx context.Context, coinTag string) (bool, error) {
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
		fmt.Println("error in unmarshal:", err)
		return false, err
	}

	result := data["result"].([]interface{})
	if len(result) == 0 {
		fmt.Println("empty result")
		return false, nil
	}

	return true, nil
}
