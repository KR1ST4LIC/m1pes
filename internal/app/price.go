package app

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log/slog"
	"m1pes/internal/models"
	"net/http"
	"strconv"
	"time"
)

const (
	priceURL    = "https://api.bybit.com/v2/public/tickers?symbol="
	apiKey      = "e6jg0dLQEagHAiBvk6"
	reqCoolDown = 1000 // in milliseconds
)

func (a *App) ParsingPrice(ctx context.Context, coinTag string) error {
	for {
		time.Sleep(reqCoolDown * time.Millisecond)

		url := priceURL + coinTag

		cli := &http.Client{
			Timeout: 5 * time.Minute,
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Error creating request:", err)
			return err
		}

		req.Header.Set("X-BAPI-API-KEY", apiKey)
		resp, err := cli.Do(req)
		if err != nil {
			slog.ErrorContext(ctx, "Error sending request:", err)
			return err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			slog.ErrorContext(ctx, "Error reading response body:", err)
			return err
		}

		var data map[string]interface{}

		if err = json.Unmarshal(body, &data); err != nil {
			slog.ErrorContext(ctx, "error in unmarshal:", err)
			return err
		}

		result := data["result"].([]interface{})
		if len(result) == 0 {
			slog.ErrorContext(ctx, "empty result")
			return nil
		}

		bidPrice := result[0].(map[string]interface{})["bid_price"].(string)
		price, _ := strconv.ParseFloat(bidPrice, 64)

		models.CoinPrice[coinTag] = price

		return nil
	}
}
