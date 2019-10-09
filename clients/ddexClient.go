package clients

import (
"encoding/json"
"errors"
"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
"github.com/sirupsen/logrus"
"strconv"
"time"
)

type DdexClient struct {
	PrivateKey string
	Address            string
	QuoteTokenSymbol   string
	QuoteTokenAddress  string
	BaseTokenSymbol    string
	BaseTokenAddress   string
	BasicUrl            string
	PricePrecision     int
	PriceDecimal       int
	AmountDecimal      int
	MinAmount          decimal.Decimal
}

func GetDdexClient(privateKey string, tradingPair string){

}

func (client *DdexClient) TradingPair() string {
	return client.BaseTokenSymbol + "-" + client.QuoteTokenSymbol
}

func (client *DdexClient) get(path string, params []utils.KeyPair) (string, error) {
	hydroAuthStr, err := privateKeyManager.PkmGetHydroAuth(client.PrivateKeyNickName)
	if err != nil {
		return "", err
	}
	return utils.Get(
		utils.JoinUrlPath(client.BaseUrl, path),
		"",
		params,
		[]utils.KeyPair{
			{"Hydro-Authentication", hydroAuthStr},
			{"Content-Type", "application/json"},
		},
	)
}

func (client *DdexClient) post(path string, body string, params []utils.KeyPair) (string, error) {
	hydroAuthStr, err := privateKeyManager.PkmGetHydroAuth(client.PrivateKeyNickName)
	if err != nil {
		return "", err
	}
	return utils.Post(
		utils.JoinUrlPath(client.BaseUrl, path),
		body,
		params,
		[]utils.KeyPair{
			{"Hydro-Authentication", hydroAuthStr},
			{"Content-Type", "application/json"},
		},
	)
}

func (client *DdexClient) delete(path string, params []utils.KeyPair) (string, error) {
	hydroAuthStr, err := privateKeyManager.PkmGetHydroAuth(client.PrivateKeyNickName)
	if err != nil {
		return "", err
	}
	return utils.Delete(
		utils.JoinUrlPath(client.BaseUrl, path),
		"",
		params,
		[]utils.KeyPair{
			{"Hydro-Authentication", hydroAuthStr},
			{"Content-Type", "application/json"},
		},
	)
}

func (client *DdexClient) Init() error {
	address, err := privateKeyManager.PkmGetAddressFromNickname(client.PrivateKeyNickName)
	if err != nil {
		return err
	}
	client.Address = address
	var dataContainer apiResponseInterface.IDdexMarkets
	resp, err := client.get("markets/"+client.TradingPair(), utils.EmptyKeyPairList)
	if err != nil {
		return err
	}
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return errors.New(fmt.Sprintf("unmarshal failed %s", resp))
	}
	client.BaseTokenAddress = dataContainer.Data.Market.BaseAssetAddress
	client.QuoteTokenAddress = dataContainer.Data.Market.QuoteAssetAddress
	client.PriceDecimal = dataContainer.Data.Market.PriceDecimals
	client.PricePrecision = dataContainer.Data.Market.PricePrecision
	client.AmountDecimal = dataContainer.Data.Market.AmountDecimals
	minAmount, err := decimal.NewFromString(dataContainer.Data.Market.MinOrderSize)
	if err != nil {
		return errors.New(fmt.Sprintf("parse min order size failed %s", dataContainer.Data.Market.MinOrderSize))
	} else {
		client.MinAmount = minAmount
	}
	return nil
}

func (client *DdexClient) buildUnsignedOrder(
	price decimal.Decimal,
	amount decimal.Decimal,
	side string,
	orderType string,
	isMakerOnly bool,
	expireTimeInSecond int64) (string, error) {
	var dataContainer apiResponseInterface.IBuildOrder
	var body = struct {
		MarketId    string          `json:"marketId"`
		Side        string          `json:"side"`
		OrderType   string          `json:"orderType"`
		Price       decimal.Decimal `json:"price"`
		Amount      decimal.Decimal `json:"amount"`
		Expires     int64           `json:"expires"`
		IsMakerOnly bool            `json:"isMakerOnly"`
		WalletType  string          `json:"walletType"`
	}{client.TradingPair(), side, orderType, price, amount, expireTimeInSecond, isMakerOnly, "trading"}
	bodyBytes, _ := json.Marshal(body)
	resp, err := client.post("orders/build", string(bodyBytes), utils.EmptyKeyPairList)
	if err != nil {
		return "", err
	}
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return "", errors.New(resp)
	} else {
		return dataContainer.Data.Order.ID, nil
	}
}

func (client *DdexClient) placeOrder(orderId string) bool {
	signature, err := privateKeyManager.PkmSignOrderId(client.PrivateKeyNickName, orderId)
	if err != nil {
		return false
	}
	var body = struct {
		OrderId   string `json:"orderId"`
		Signature string `json:"signature"`
	}{orderId, signature}
	bodyBytes, _ := json.Marshal(body)
	resp, err := client.post("orders", string(bodyBytes), utils.EmptyKeyPairList)
	if err != nil {
		return false
	}
	var dataContainer apiResponseInterface.IPlaceOrder
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		spew.Dump(dataContainer)
		return false
	} else {
		return true
	}
}

func (client *DdexClient) placeOrderSynchronously(orderId string) (*StdOrder, error) {
	signature, err := privateKeyManager.PkmSignOrderId(client.PrivateKeyNickName, orderId)
	if err != nil {
		return nil, err
	}
	var body = struct {
		OrderId   string `json:"orderId"`
		Signature string `json:"signature"`
	}{orderId, signature}
	bodyBytes, _ := json.Marshal(body)
	resp, err := client.post("orders/sync", string(bodyBytes), utils.EmptyKeyPairList)
	if err != nil {
		return nil, err
	}
	var dataContainer apiResponseInterface.IPlaceOrderSync
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return nil, errors.New("")
	} else {
		orderInfo := client.parseDdexOrderResp(dataContainer.Data.Order)
		return &orderInfo, nil
	}
}

func (client *DdexClient) CreateMarketOrder(
	priceLimit decimal.Decimal,
	amount decimal.Decimal,
	side string,
) (*StdOrder, error) {

	validPrice := utils.SetDecimal(utils.SetPrecision(priceLimit, client.PricePrecision), client.PriceDecimal)
	if side == utils.SELL {
		amount = utils.SetDecimal(amount, client.AmountDecimal)
	}

	orderId, err := client.buildUnsignedOrder(validPrice, amount, side, utils.MARKET, false, 3600)
	if err != nil {
		return nil, err
	}
	orderInfo, err := client.placeOrderSynchronously(orderId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ddex client %s place order sync failed", client.TradingPair()))
	}
	if orderInfo.FilledAmount.Equal(decimal.Zero) {
		return nil, errors.New(fmt.Sprintf("ddex client %s market order price protect triggered", client.TradingPair()))
	} else {
		logrus.Infof("ddex client %s create market order - price:%s amount:%s side:%s %s", client.TradingPair(), validPrice, amount, side, orderId)
		return orderInfo, nil
	}
}

func (client *DdexClient) CreateLimitOrder(
	price decimal.Decimal,
	amount decimal.Decimal,
	side string,
	isMakerOnly bool,
	expireTimeInSecond int64) (string, error) {

	validPrice := utils.SetDecimal(utils.SetPrecision(price, client.PricePrecision), client.PriceDecimal)

	validAmount := utils.SetDecimal(amount, client.AmountDecimal)
	if validAmount.Mul(validPrice).LessThan(client.MinAmount) {
		return "", errors.New(fmt.Sprintf("ddex client %s create order amount %s less than min amount %s", client.TradingPair(), validAmount.String(), client.MinAmount.String()))
	}

	orderId, err := client.buildUnsignedOrder(validPrice, validAmount, side, utils.LIMIT, isMakerOnly, expireTimeInSecond)
	if err != nil {
		return "", err
	}
	placeSuccess := client.placeOrder(orderId)
	if placeSuccess {
		logrus.Infof("ddex client %s create limit order - price:%s amount:%s side:%s %s", client.TradingPair(), validPrice, validAmount, side, orderId)
		return orderId, nil
	} else {
		return "", errors.New(fmt.Sprintf("ddex client %s place order failed", client.TradingPair()))
	}
}

func (client *DdexClient) CancelOrder(orderId string) error {
	resp, err := client.delete("orders/"+orderId, utils.EmptyKeyPairList)
	if err != nil {
		return err
	}
	var dataContainer apiResponseInterface.ICancelOrder
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return errors.New(fmt.Sprintf("ddex client %s cancel order %s failed", client.TradingPair(), orderId))
	} else {
		logrus.Infof("ddex client %s cancel order %s succeed", client.TradingPair(), orderId)
		return nil
	}
}

func (client *DdexClient) PromisedCancelOrder(orderId string) {
	for true {
		err := client.CancelOrder(orderId)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func (client *DdexClient) CancelAllPendingOrders() (bool, error) {
	resp, err := client.delete("orders", []utils.KeyPair{{"marketId", client.TradingPair()}})
	if err != nil {
		return false, err
	}
	var dataContainer apiResponseInterface.ICancelOrder
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return false, errors.New(fmt.Sprintf("ddex client %s cancel all pending orders failed", client.TradingPair()))
	} else {
		logrus.Infof("ddex client %s cancel all orders succeed", client.TradingPair())
		return true, nil
	}
}

func (client *DdexClient) parseDdexOrderResp(orderInfo apiResponseInterface.IDDEXOrderResp) StdOrder {
	var orderData = EmptyStdOrder
	orderData.Id = orderInfo.ID
	orderData.Amount, _ = decimal.NewFromString(orderInfo.Amount)
	orderData.AvailableAmount, _ = decimal.NewFromString(orderInfo.AvailableAmount)
	orderData.Price, _ = decimal.NewFromString(orderInfo.Price)
	orderData.AvgPrice, _ = decimal.NewFromString(orderInfo.AveragePrice)
	pendingAmount, _ := decimal.NewFromString(orderInfo.PendingAmount)
	confirmedAmount, _ := decimal.NewFromString(orderInfo.ConfirmedAmount)
	orderData.FilledAmount = pendingAmount.Add(confirmedAmount)
	if orderData.AvailableAmount.IsZero() {
		orderData.Status = utils.ORDER_CLOSE
	} else {
		orderData.Status = utils.ORDER_OPEN
	}
	if orderInfo.Side == "sell" {
		orderData.Side = utils.SELL
	} else {
		orderData.Side = utils.BUY
	}
	return orderData
}

func (client *DdexClient) GetOrder(orderId string) (StdOrder, error) {
	orderData := EmptyStdOrder
	resp, err := client.get("orders/"+orderId, utils.EmptyKeyPairList)
	if err != nil {
		return orderData, err
	}
	var dataContainer apiResponseInterface.IOrder
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return orderData, errors.New(fmt.Sprintf("ddex client %s get order failed", client.TradingPair()))
	} else {
		orderData = client.parseDdexOrderResp(dataContainer.Data.Order)
		return orderData, nil
	}
}

func (client *DdexClient) PromisedGetClosedOrder(orderId string) StdOrder {
	var orderInfo StdOrder
	for true {
		info, err := client.GetOrder(orderId)
		if err == nil && info.Status == utils.ORDER_CLOSE {
			orderInfo = info
			break
		}
		time.Sleep(1 * time.Second)
	}
	return orderInfo
}

func (client *DdexClient) GetAllPendingOrders() ([]StdOrder, error) {
	var allOrders = []StdOrder{}
	var pageNum = 1
	for true {
		resp, err := client.get("orders", []utils.KeyPair{
			{"marketId", client.TradingPair()},
			{"perPage", "100"},
			{"page", strconv.Itoa(pageNum)},
		})
		if err != nil {
			return allOrders, err
		}
		var dataContainer apiResponseInterface.IAllPendingOrders
		json.Unmarshal([]byte(resp), &dataContainer)
		if dataContainer.Desc != "success" {
			return allOrders, errors.New(fmt.Sprintf("ddex client %s get all pending orders failed", client.TradingPair()))
		}
		for _, order := range dataContainer.Data.Orders {
			var tempOrder = client.parseDdexOrderResp(order)
			allOrders = append(allOrders, tempOrder)
		}
		if pageNum >= dataContainer.Data.TotalPages {
			break
		}
		pageNum += 1
	}
	return allOrders, nil
}

func (client *DdexClient) GetInventory() (inventory Inventory, err error) {
	inventory = EmptyInventory

	_, totalBaseAmount, err := utils.GetBfdBalance(client.BaseTokenAddress, client.Address)
	if err != nil {
		return
	}
	_, totalQuoteAmount, err := utils.GetBfdBalance(client.QuoteTokenAddress, client.Address)
	if err != nil {
		return
	}
	totalBaseAmount, err = utils.ConvertToReadableValue(client.BaseTokenAddress, totalBaseAmount)
	if err != nil {
		return
	}
	totalQuoteAmount, err = utils.ConvertToReadableValue(client.QuoteTokenAddress, totalQuoteAmount)
	if err != nil {
		return
	}

	resp, err := client.get("account/lockedBalances", utils.EmptyKeyPairList)
	if err != nil {
		return
	}
	var dataContainer apiResponseInterface.ILockedBalance
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return inventory, err
	}
	lockedBase := &decimal.Zero
	lockedQuote := &decimal.Zero
	for _, lockedBalance := range dataContainer.Data.LockedBalances {
		if lockedBalance.AssetAddress == client.BaseTokenAddress && lockedBalance.WalletType == "trading" {
			amount, err := decimal.NewFromString(lockedBalance.Amount)
			if err != nil {
				return inventory, err
			}
			readableAmount, err := utils.ConvertToReadableValue(client.BaseTokenAddress, &amount)
			if err != nil {
				return inventory, err
			}
			lockedBase = readableAmount
		}
		if lockedBalance.AssetAddress == client.QuoteTokenAddress && lockedBalance.WalletType == "trading" {
			amount, err := decimal.NewFromString(lockedBalance.Amount)
			if err != nil {
				return inventory, err
			}
			readableAmount, err := utils.ConvertToReadableValue(client.QuoteTokenAddress, &amount)
			if err != nil {
				return inventory, err
			}
			lockedQuote = readableAmount
		}
	}

	inventory.Base.Total = *totalBaseAmount
	inventory.Base.Lock = *lockedBase
	inventory.Base.Free = totalBaseAmount.Sub(*lockedBase)

	inventory.Quote.Total = *totalQuoteAmount
	inventory.Quote.Lock = *lockedQuote
	inventory.Quote.Free = totalQuoteAmount.Sub(*lockedQuote)

	return inventory, nil
}

func (client *DdexClient) GetTicker() (Ticker, error) {
	var ticker = EmptyTicker
	resp, err := client.get("markets/"+client.TradingPair()+"/ticker", utils.EmptyKeyPairList)
	if err != nil {
		return ticker, err
	}
	var dataContainer apiResponseInterface.ITicker
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return ticker, errors.New(fmt.Sprintf("ddex client %s get ticker failed", client.TradingPair()))
	}
	ticker.LastPrice, _ = decimal.NewFromString(dataContainer.Data.Ticker.Price)
	ticker.BuyPrice, _ = decimal.NewFromString(dataContainer.Data.Ticker.Bid)
	ticker.SellPrice, _ = decimal.NewFromString(dataContainer.Data.Ticker.Ask)
	return ticker, nil
}

func (client *DdexClient) GetCurrentPrice() (
	bestBidPrice decimal.Decimal,
	bestAskPrice decimal.Decimal,
	centerPrice decimal.Decimal,
	err error) {
	resp, err := client.get(
		fmt.Sprintf("markets/%s/orderbook", client.TradingPair()),
		[]utils.KeyPair{{"level", "1"}},
	)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	var dataContainer apiResponseInterface.ILevel1Orderbook
	json.Unmarshal([]byte(resp), &dataContainer)
	if len(dataContainer.Data.OrderBook.Asks) == 0 || len(dataContainer.Data.OrderBook.Bids) == 0 {
		return decimal.Zero, decimal.Zero, decimal.Zero, errors.New("ddex client get current price failed")
	}
	bestAskPrice, _ = decimal.NewFromString(dataContainer.Data.OrderBook.Asks[0].Price)
	bestBidPrice, _ = decimal.NewFromString(dataContainer.Data.OrderBook.Bids[0].Price)
	centerPrice = bestAskPrice.Add(bestBidPrice).Div(decimal.New(2, 0))

	return bestBidPrice, bestAskPrice, centerPrice, nil
}
