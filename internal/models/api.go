package models

type GetCoinRequest struct {
	Category string `json:"category"`
	Symbol   string `json:"symbol"`
}

type GetCoinResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		Category string `json:"category"`
		List     []struct {
			Symbol string `json:"symbol"`
			Price  string `json:"lastPrice"`
		} `json:"list"`
	} `json:"result"`
}

type CreateOrderRequest struct {
	Category    string `json:"category"`
	Side        string `json:"side"`
	Symbol      string `json:"symbol"`
	OrderType   string `json:"orderType"`
	Qty         string `json:"qty"`
	MarketUint  string `json:"marketUint"`
	PositionIdx int    `json:"positionIdx"`
	Price       string `json:"price"`
	TimeInForce string `json:"timeInForce"`
}

type CreateOrderResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	ExtCode string `json:"extCode"`
	ExtInfo string `json:"extInfo"`
	Result  struct {
		OrderID     string `json:"orderId"`
		OrderLinkID string `json:"orderLinkId"`
	} `json:"result"`
	TimeNow string `json:"timeNow"`
}

type CancelOrderRequest struct {
	Category string `json:"category"`
	OrderId  string `json:"orderId"`
	Symbol   string `json:"symbol"`
}

type CancelOrderResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		OrderID     string `json:"orderId"`
		OrderLinkID string `json:"orderLinkId"`
	} `json:"result"`
	TimeNow string `json:"timeNow"`
}

type GetOrderRequest struct {
	Category string `json:"category"`
	OrderID  string `json:"orderId"`
	Symbol   string `json:"symbol"`
}

type GetOrderResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	List    []struct {
		OrderStatus string `json:"orderStatus"`
		Price       string `json:"price"`
		Qty         string `json:"qty"`
		Side        string `json:"side"`
	} `json:"list"`
}

type OrderCreate struct {
	Symbol string
	Side   string
	Qty    string
	Price  string
}

type Response struct {
	RetCode    int      `json:"retCode"`
	RetMsg     string   `json:"retMsg"`
	Result     Result   `json:"result"`
	RetExtInfo struct{} `json:"retExtInfo"`
	Time       int64    `json:"time"`
}

type Result struct {
	List []Account `json:"list"`
}

type Account struct {
	TotalEquity        string `json:"totalEquity"`
	TotalWalletBalance string `json:"totalWalletBalance"`
}
