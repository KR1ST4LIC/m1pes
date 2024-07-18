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
	"m1pes/internal/logging"
	"m1pes/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	URL = "https://api.bybit.com"

	CreateOrderEndpoint   = "/v5/order/create"
	CancelOrderEndpoint   = "/v5/order/cancel"
	GetOrderEndpoint      = "/v5/order/realtime"
	GetUserWalletEndpoint = "/v5/account/wallet-balance"
	GetCoinEndpoint       = "/v5/market/tickers"
	GetApiKeyPermissions  = "/v5/user/query-api"

	SuccessfulOrderStatus = "Filled"
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

func (r *Repository) GetCoin(ctx context.Context, coinReq models.GetCoinRequest, apiKey, secretKey string) (models.GetCoinResponse, error) {
	byteParams, err := json.Marshal(coinReq)
	if err != nil {
		return models.GetCoinResponse{}, fmt.Errorf("marshal get coin request failed: %w", err)
	}

	var getCoinResp models.GetCoinResponse
	body, err := r.CreateSignRequestAndGetRespBody(string(byteParams), GetCoinEndpoint, http.MethodGet, apiKey, secretKey)
	if err != nil {
		return models.GetCoinResponse{}, fmt.Errorf("get coin request failed: %w", err)
	}

	err = json.Unmarshal(body, &getCoinResp)
	if err != nil {
		return models.GetCoinResponse{}, fmt.Errorf("unmarshal get coin response failed: %w", err)
	}

	if getCoinResp.RetMsg != "OK" {
		return models.GetCoinResponse{}, fmt.Errorf("get coin response failed: %s", getCoinResp.RetMsg)
	}

	return getCoinResp, nil
}

func (r *Repository) GetUserWalletBalance(ctx context.Context, req models.GetUserWalletRequest, apiKey, secretKey string) (models.GetUserWalletResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return models.GetUserWalletResponse{}, fmt.Errorf("marshal get user's wallet request failed: %w", err)
	}

	body, err := r.CreateSignRequestAndGetRespBody(string(jsonData), GetUserWalletEndpoint, http.MethodGet, apiKey, secretKey)
	if err != nil {
		return models.GetUserWalletResponse{}, fmt.Errorf("get user's wallet request failed: %w", err)
	}

	var getUserWalletResp models.GetUserWalletResponse

	err = json.Unmarshal(body, &getUserWalletResp)
	if err != nil {
		return models.GetUserWalletResponse{}, fmt.Errorf("unmarshal get user's wallet response failed: %w", err)
	}

	if getUserWalletResp.RetMsg != "OK" {
		return models.GetUserWalletResponse{}, fmt.Errorf("get user's wallet failed: %s", getUserWalletResp.RetMsg)
	}

	return getUserWalletResp, nil
}

func (r *Repository) CreateOrder(ctx context.Context, orderReq models.CreateOrderRequest, apiKey, secretKey string) (models.CreateOrderResponse, error) {
	jsonData, err := json.Marshal(orderReq)
	if err != nil {
		return models.CreateOrderResponse{}, logging.WrapError(ctx, fmt.Errorf("marshal order request failed: %w", err))
	}

	body, err := r.CreateSignRequestAndGetRespBody(string(jsonData), CreateOrderEndpoint, http.MethodPost, apiKey, secretKey)
	if err != nil {
		return models.CreateOrderResponse{}, fmt.Errorf("create order request failed: %w", err)
	}

	var createOrderResp models.CreateOrderResponse
	err = json.Unmarshal(body, &createOrderResp)
	if err != nil {
		return models.CreateOrderResponse{}, fmt.Errorf("unmarshal create order response failed: %w", err)
	}

	if createOrderResp.RetMsg != "OK" {
		return models.CreateOrderResponse{}, fmt.Errorf("create order failed: %s", createOrderResp.RetMsg)
	}

	return createOrderResp, nil
}

func (r *Repository) CancelOrder(ctx context.Context, orderReq models.CancelOrderRequest, apiKey, secretKey string) (models.CancelOrderResponse, error) {
	jsonData, err := json.Marshal(orderReq)
	if err != nil {
		return models.CancelOrderResponse{}, fmt.Errorf("marshal cancel order request failed: %w", err)
	}

	body, err := r.CreateSignRequestAndGetRespBody(string(jsonData), CancelOrderEndpoint, http.MethodPost, apiKey, secretKey)
	if err != nil {
		return models.CancelOrderResponse{}, fmt.Errorf("create cancel order request failed: %w", err)
	}

	var cancelOrderResp models.CancelOrderResponse
	err = json.Unmarshal(body, &cancelOrderResp)
	if err != nil {
		return models.CancelOrderResponse{}, fmt.Errorf("unmarshal cancel order response failed: %w", err)
	}

	if cancelOrderResp.RetMsg != "OK" {
		return models.CancelOrderResponse{}, fmt.Errorf("cancel order failed: %s", cancelOrderResp.RetMsg)
	}

	return cancelOrderResp, nil
}

func (r *Repository) GetOrder(ctx context.Context, orderReq models.GetOrderRequest, apiKey, secretKey string) (models.GetOrderResponse, error) {
	jsonData, err := json.Marshal(orderReq)
	if err != nil {
		return models.GetOrderResponse{}, fmt.Errorf("marshalling order request failed: %w", err)
	}

	body, err := r.CreateSignRequestAndGetRespBody(string(jsonData), GetOrderEndpoint, http.MethodGet, apiKey, secretKey)
	if err != nil {
		return models.GetOrderResponse{}, fmt.Errorf("create order request failed: %w", err)
	}

	var getOrderResp models.GetOrderResponse

	err = json.Unmarshal(body, &getOrderResp)
	if err != nil {
		return models.GetOrderResponse{}, fmt.Errorf("unmarshal order response failed: %w", err)
	}

	if getOrderResp.RetMsg != "OK" {
		return models.GetOrderResponse{}, fmt.Errorf("get order failed: %s", getOrderResp.RetMsg)
	}

	return getOrderResp, nil
}

func (r *Repository) CreateSignRequestAndGetRespBody(params, endPoint, method, apiKey, apiSecret string) ([]byte, error) {
	var request *http.Request
	switch method {
	case http.MethodGet:
		if params != "" {
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
		} else {
			req, err := http.NewRequest(method, URL+endPoint, nil)
			if err != nil {
				return nil, errors.Wrap(err, "failed create new request")
			}
			request = req
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
