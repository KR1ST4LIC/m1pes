package models

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
