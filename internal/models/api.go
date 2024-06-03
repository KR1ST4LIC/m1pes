package models

type OrderCreate struct {
	Symbol string
	Side   string
	Qty    string
	Price  string
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
