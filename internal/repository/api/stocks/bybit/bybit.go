package bybit

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	priceURL = "https://api.bybit.com/v2/public/tickers?symbol="
	URL      = "https://api.bybit.com"
)

type Repository struct {
	cli *http.Client
}

func New() *Repository {
	return &Repository{
		cli: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (r *Repository) GetPrice(ctx context.Context, coinTag, apiKey string) (float64, error) {
	url := priceURL + coinTag

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error creating request:", err)
		return 0, err
	}

	req.Header.Set("X-BAPI-API-KEY", apiKey)
	resp, err := r.cli.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "Error sending request:", err)
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Error reading response body:", err)
		return 0, err
	}

	var data map[string]interface{}

	if err = json.Unmarshal(body, &data); err != nil {
		slog.ErrorContext(ctx, "error in unmarshal:", err)
		return 0, err
	}

	result := data["result"].([]interface{})
	if len(result) == 0 {
		slog.ErrorContext(ctx, "empty result")
		return 0, nil
	}

	bidPrice := result[0].(map[string]interface{})["bid_price"].(string)

	price, _ := strconv.ParseFloat(bidPrice, 64)

	return price, nil
}

func (r *Repository) ExistCoin(ctx context.Context, coinTag, apiKey string) (bool, error) {
	url := priceURL + coinTag

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error creating request:", err)
		return false, err
	}

	req.Header.Set("X-BAPI-API-KEY", apiKey)
	resp, err := r.cli.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "Error sending request:", err)
		return false, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Error reading response body:", err)
		return false, err
	}

	var data map[string]interface{}

	if err = json.Unmarshal(body, &data); err != nil {
		slog.ErrorContext(ctx, "error in unmarshal:", err)
		return false, err
	}

	result := data["result"].([]interface{})
	if len(result) == 0 {
		slog.ErrorContext(ctx, "empty result")
		return false, nil
	}

	return true, nil
}

func (r *Repository) CreateSignRequestAndGetRespBody(params, endPoint, method, apiKey, apiSecret string) ([]byte, error) {
	var request *http.Request
	switch method {
	case http.MethodGet:
		paramsMap := make(map[string]interface{})

		err := json.Unmarshal([]byte(params), &paramsMap)
		if err != nil {
			return nil, err
		}

		params = ""
		isFirst := true
		for key, val := range paramsMap {
			if isFirst {
				params += fmt.Sprintf("%s=%v", key, val)
				isFirst = false
				continue
			}
			params += fmt.Sprintf("&%s=%v", key, val)
		}

		request, err = http.NewRequest(method, URL+endPoint+"?"+params, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed create new request")
		}
	case http.MethodPost:
		newRequest, err := http.NewRequest(method, URL+endPoint, bytes.NewBuffer([]byte(params)))
		if err != nil {
			return nil, errors.Wrap(err, "failed create new request")
		}

		request = newRequest
	default:
		return nil, errors.Errorf("unsupported method: %s", method)
	}

	timestamp := time.Now().UnixMilli()
	hmac256 := hmac.New(sha256.New, []byte(apiSecret))
	hmac256.Write([]byte(strconv.FormatInt(timestamp, 10) + apiKey + "5000" + params))
	signature := hex.EncodeToString(hmac256.Sum(nil))

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-BAPI-API-KEY", apiKey)
	request.Header.Set("X-BAPI-SIGN", signature)
	request.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timestamp, 10))
	request.Header.Set("X-BAPI-SIGN-TYPE", "2")
	request.Header.Set("X-BAPI-RECV-WINDOW", "5000")

	resp, err := r.cli.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed do request")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed read body")
	}

	return data, err
}
