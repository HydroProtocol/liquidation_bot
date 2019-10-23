package client

type IMarkets struct {
	Status   int         `json:"status"`
	Desc     string      `json:"desc"`
	Template string      `json:"template"`
	Params   interface{} `json:"params"`
	Data     struct {
		Markets []struct {
			ID                        string   `json:"id"`
			MaxSlippage               string   `json:"maxSlippage"`
			MarginMarketID            int      `json:"marginMarketId"`
			IsMarginMarket            bool     `json:"isMarginMarket"`
			BaseAsset                 string   `json:"baseAsset"`
			BaseAssetName             string   `json:"baseAssetName"`
			BaseAssetDecimals         int      `json:"baseAssetDecimals"`
			BaseAssetDisplayDecimals  int      `json:"baseAssetDisplayDecimals"`
			BaseAssetAddress          string   `json:"baseAssetAddress"`
			QuoteAsset                string   `json:"quoteAsset"`
			QuoteAssetName            string   `json:"quoteAssetName"`
			QuoteAssetDecimals        int      `json:"quoteAssetDecimals"`
			QuoteAssetDisplayDecimals int      `json:"quoteAssetDisplayDecimals"`
			QuoteAssetAddress         string   `json:"quoteAssetAddress"`
			MinOrderSize              string   `json:"minOrderSize"`
			PricePrecision            int      `json:"pricePrecision"`
			PriceDecimals             int      `json:"priceDecimals"`
			AmountDecimals            int      `json:"amountDecimals"`
			AsMakerFeeRate            string   `json:"asMakerFeeRate"`
			AsTakerFeeRate            string   `json:"asTakerFeeRate"`
			GasFeeAmount              string   `json:"gasFeeAmount"`
			SupportedOrderTypes       []string `json:"supportedOrderTypes"`
			LastPriceIncrease         string   `json:"lastPriceIncrease"`
			LastPrice                 string   `json:"lastPrice"`
			Price24H                  string   `json:"price24h"`
			Amount24H                 string   `json:"amount24h"`
			QuoteAssetVolume24H       string   `json:"quoteAssetVolume24h"`
			BaseAssetUSDPrice         string   `json:"baseAssetUSDPrice"`
			QuoteAssetUSDPrice        string   `json:"quoteAssetUSDPrice"`
			MaxLeverageRate           string   `json:"maxLeverageRate"`
			WithdrawRate              string   `json:"withdrawRate"`
			LiquidationRate           string   `json:"liquidationRate"`
		} `json:"markets"`
	} `json:"data"`
}

type IAssets struct {
	Status   int         `json:"status"`
	Desc     string      `json:"desc"`
	Template string      `json:"template"`
	Params   interface{} `json:"params"`
	Data     struct {
		Assets []struct {
			Address            string `json:"address"`
			DisplayDecimals    int    `json:"displayDecimals"`
			Symbol             string `json:"symbol"`
			Name               string `json:"name"`
			Decimals           int    `json:"decimals"`
			OracleUSDPrice     string `json:"oracleUSDPrice"`
			BorrowInterestRate string `json:"borrowInterestRate"`
			SupplyInterestRate string `json:"supplyInterestRate"`
			PoolTokenAddress   string `json:"poolTokenAddress"`
		} `json:"assets"`
	} `json:"data"`
}

type IOrderResp struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Version         string `json:"version"`
	Status          string `json:"status"`
	Amount          string `json:"amount"`
	AvailableAmount string `json:"availableAmount"`
	PendingAmount   string `json:"pendingAmount"`
	CanceledAmount  string `json:"canceledAmount"`
	ConfirmedAmount string `json:"confirmedAmount"`
	Price           string `json:"price"`
	AveragePrice    string `json:"averagePrice"`
	Side            string `json:"side"`
	MakerFeeRate    string `json:"makerFeeRate"`
	TakerFeeRate    string `json:"takerFeeRate"`
	MakerRebateRate string `json:"makerRebateRate"`
	GasFeeAmount    string `json:"gasFeeAmount"`
	Account         string `json:"account"`
	CreatedAt       int64  `json:"createdAt"`
	MarketID        string `json:"marketId"`
}

type ILockedBalance struct {
	Status int    `json:"status"`
	Desc   string `json:"desc"`
	Data   struct {
		LockedBalances []struct {
			Address        string `json:"Address"`
			Symbol         string `json:"symbol"`
			AssetAddress   string `json:"assetAddress"`
			WalletType     string `json:"walletType"`
			MarginMarketID int    `json:"marginMarketId"`
			Amount         string `json:"amount"`
		} `json:"lockedBalances"`
	} `json:"data"`
}

type IOrder struct {
	Status int    `json:"status"`
	Desc   string `json:"desc"`
	Data   struct {
		Order IOrderResp `json:"order"`
	} `json:"data"`
}

type IAllPendingOrders struct {
	Status int    `json:"status"`
	Desc   string `json:"desc"`
	Data   struct {
		TotalCount  int          `json:"totalCount"`
		TotalPages  int          `json:"totalPages"`
		CurrentPage int          `json:"currentPage"`
		Orders      []IOrderResp `json:"orders"`
	} `json:"data"`
}

type IBuildOrder struct {
	Status int    `json:"status"`
	Desc   string `json:"desc"`
	Data   struct {
		Order struct {
			ID string `json:"id"`
		} `json:"order"`
	} `json:"data"`
}

type IPlaceOrder struct {
	Status int    `json:"status"`
	Desc   string `json:"desc"`
}

type IPlaceOrderSync struct {
	Status int    `json:"status"`
	Desc   string `json:"desc"`
	Data   struct {
		Order IOrderResp `json:"order"`
	} `json:"data"`
}

type ICancelOrder struct {
	Status int    `json:"status"`
	Desc   string `json:"desc"`
}

type IOrderbook struct {
	Status int    `json:"status"`
	Desc   string `json:"desc"`
	Data   struct {
		OrderBook struct {
			MarketID string `json:"marketId"`
			Bids     []struct {
				Price  string `json:"price"`
				Amount string `json:"amount"`
			} `json:"bids"`
			Asks []struct {
				Price  string `json:"price"`
				Amount string `json:"amount"`
			} `json:"asks"`
		} `json:"orderBook"`
	} `json:"data"`
}
