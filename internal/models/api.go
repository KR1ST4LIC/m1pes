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
		NextPageCursor string `json:"nextPageCursor"`
		Category       string `json:"category"`
		List           []struct {
			Symbol             string `json:"symbol"`
			OrderType          string `json:"orderType"`
			OrderLinkId        string `json:"orderLinkId"`
			SlLimitPrice       string `json:"slLimitPrice"`
			OrderId            string `json:"orderId"`
			CancelType         string `json:"cancelType"`
			AvgPrice           string `json:"avgPrice"`
			StopOrderType      string `json:"stopOrderType"`
			LastPriceOnCreated string `json:"lastPriceOnCreated"`
			OrderStatus        string `json:"orderStatus"`
			TakeProfit         string `json:"takeProfit"`
			CumExecValue       string `json:"cumExecValue"`
			SmpType            string `json:"smpType"`
			TriggerDirection   int    `json:"triggerDirection"`
			BlockTradeId       string `json:"blockTradeId"`
			IsLeverage         string `json:"isLeverage"`
			RejectReason       string `json:"rejectReason"`
			Price              string `json:"price"`
			OrderIv            string `json:"orderIv"`
			CreatedTime        string `json:"createdTime"`
			TpTriggerBy        string `json:"tpTriggerBy"`
			PositionIdx        int    `json:"positionIdx"`
			TrailingPercentage string `json:"trailingPercentage"`
			TimeInForce        string `json:"timeInForce"`
			LeavesValue        string `json:"leavesValue"`
			BasePrice          string `json:"basePrice"`
			UpdatedTime        string `json:"updatedTime"`
			Side               string `json:"side"`
			SmpGroup           int    `json:"smpGroup"`
			TriggerPrice       string `json:"triggerPrice"`
			TpLimitPrice       string `json:"tpLimitPrice"`
			TrailingValue      string `json:"trailingValue"`
			CumExecFee         string `json:"cumExecFee"`
			LeavesQty          string `json:"leavesQty"`
			SlTriggerBy        string `json:"slTriggerBy"`
			CloseOnTrigger     bool   `json:"closeOnTrigger"`
			PlaceType          string `json:"placeType"`
			CumExecQty         string `json:"cumExecQty"`
			ReduceOnly         bool   `json:"reduceOnly"`
			ActivationPrice    string `json:"activationPrice"`
			Qty                string `json:"qty"`
			StopLoss           string `json:"stopLoss"`
			MarketUnit         string `json:"marketUnit"`
			SmpOrderId         string `json:"smpOrderId"`
			TriggerBy          string `json:"triggerBy"`
		} `json:"list"`
	} `json:"result"`
	RetExtInfo struct {
	} `json:"retExtInfo"`
	Time int64 `json:"time"`
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
