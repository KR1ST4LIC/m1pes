package models

// -----Get coin endpoint------

type GetCoinRequest map[string]interface{}

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

// -----Create order endpoint------

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

// -----Cancel order endpoint------

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

// -----Get order endpoint------

type GetOrderRequest map[string]interface{}

type GetOrderResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		List []struct {
			OrderID     string `json:"orderId"`
			OrderStatus string `json:"orderStatus"`
			Price       string `json:"price"`
			Qty         string `json:"qty"`
			Side        string `json:"side"`
		} `json:"list"`
	} `json:"result"`
}

// -----Get user's wallet balance endpoint------

type GetUserWalletRequest map[string]interface{}

type GetUserWalletResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		List []struct {
			//TotalWalletBalance    string `json:"totalWalletBalance"`
			//TotalAvailableBalance string `json:"totalAvailableBalance"`
			//TotalMarginBalance    string `json:"totalMarginBalance"`
			TotalEquity string `json:"totalEquity"`
			Coin        []struct {
				Coin   string `json:"coin"`
				Equity string `json:"equity"`
				//SpotHedgingQty string `json:"spotHedgingQty"`
			} `json:"coin"`
		} `json:"list"`
	} `json:"result"`
}

// -----Stupid coin endpoint------

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
