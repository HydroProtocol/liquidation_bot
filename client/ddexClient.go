package client

import (
	"auctionBidder/utils"
	"auctionBidder/web3"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"time"
)

type SimpleOrder struct {
	Amount decimal.Decimal
	Price  decimal.Decimal
	Side   string
}

type OrderRes struct {
	Id              string
	Status          string
	Side            string
	Amount          decimal.Decimal
	Price           decimal.Decimal
	AvailableAmount decimal.Decimal
	FilledAmount    decimal.Decimal
	AvgPrice        decimal.Decimal
}

type Balance struct {
	Free  decimal.Decimal
	Lock  decimal.Decimal
	Total decimal.Decimal
}

type Inventory map[string]*Balance // symbol -> Balance

type Asset struct {
	Symbol  string
	Address string
	Decimal int32
}

type Market struct {
	Base           Asset
	Quote          Asset
	PricePrecision int
	PriceDecimal   int
	AmountDecimal  int
}

type DdexClient struct {
	Address       string
	Assets        map[string]*Asset  // symbol -> Asset
	Markets       map[string]*Market // "ETH-DAI" -> Market
	hydroContract *web3.Contract
	privateKey    string
	signCache     string
	lastSignTime  int64
	baseUrl       string
}

func NewDdexClient(privateKey string) (client *DdexClient, err error) {
	ethereumNodeUrl := os.Getenv("ETHEREUM_NODE_URL")
	ddexBaseUrl := os.Getenv("DDEX_URL")
	hydroContractAddress := os.Getenv("HYDRO_CONTRACT_ADDRESS")

	web3 := web3.NewWeb3(ethereumNodeUrl)
	address, err := web3.AddPrivateKey(privateKey)
	if err != nil {
		return
	}
	contract, err := web3.NewContract(utils.HydroAbi, hydroContractAddress)
	if err != nil {
		return
	}

	// get market meta data
	var dataContainer IMarkets
	resp, err := utils.Get(
		utils.JoinUrlPath(ddexBaseUrl, fmt.Sprintf("markets")),
		"",
		utils.EmptyKeyPairList,
		[]utils.KeyPair{{"Content-Type", "application/json"}})
	if err != nil {
		return
	}
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		err = errors.New(dataContainer.Desc)
		return
	}

	assets := map[string]*Asset{}
	markets := map[string]*Market{}

	for _, market := range dataContainer.Data.Markets {
		if _, ok := assets[market.BaseAssetName]; !ok {
			assets[market.BaseAssetName] = &Asset{market.BaseAssetName, market.BaseAssetAddress, int32(market.BaseAssetDecimals)}
		}
		if _, ok := assets[market.QuoteAssetName]; !ok {
			assets[market.QuoteAssetName] = &Asset{market.QuoteAssetName, market.QuoteAssetAddress, int32(market.QuoteAssetDecimals)}
		}
		markets[fmt.Sprintf("%s-%s", market.BaseAssetName, market.QuoteAssetName)] = &Market{
			*assets[market.BaseAssetName],
			*assets[market.QuoteAssetName],
			market.PricePrecision,
			market.PriceDecimals,
			market.AmountDecimals,
		}
	}

	client = &DdexClient{
		address,
		assets,
		markets,
		contract,
		privateKey,
		"",
		0,
		ddexBaseUrl,
	}

	return
}

func (client *DdexClient) updateSignCache() {
	now := utils.MillisecondTimestamp()
	if client.lastSignTime < now-200000 {
		messageStr := "HYDRO-AUTHENTICATION@" + strconv.Itoa(int(now))
		signRes, _ := utils.PersonalSign([]byte(messageStr), client.privateKey)
		client.signCache = fmt.Sprintf("%s#%s#0x%x", strings.ToLower(client.Address), messageStr, signRes)
		client.lastSignTime = now
	}
}

func (client *DdexClient) signOrderId(orderId string) string {
	orderIdBytes, _ := hex.DecodeString(strings.TrimPrefix(orderId, "0x"))
	signature, _ := utils.PersonalSign(orderIdBytes, client.privateKey)
	return "0x" + hex.EncodeToString(signature)
}

func (client *DdexClient) get(path string, params []utils.KeyPair) (string, error) {
	client.updateSignCache()
	return utils.Get(
		utils.JoinUrlPath(client.baseUrl, path),
		"",
		params,
		[]utils.KeyPair{
			{"Hydro-Authentication", client.signCache},
			{"Content-Type", "application/json"},
		},
	)
}

func (client *DdexClient) post(path string, body string, params []utils.KeyPair) (string, error) {
	client.updateSignCache()
	return utils.Post(
		utils.JoinUrlPath(client.baseUrl, path),
		body,
		params,
		[]utils.KeyPair{
			{"Hydro-Authentication", client.signCache},
			{"Content-Type", "application/json"},
		},
	)
}

func (client *DdexClient) delete(path string, params []utils.KeyPair) (string, error) {
	client.updateSignCache()
	return utils.Delete(
		utils.JoinUrlPath(client.baseUrl, path),
		"",
		params,
		[]utils.KeyPair{
			{"Hydro-Authentication", client.signCache},
			{"Content-Type", "application/json"},
		},
	)
}

func (client *DdexClient) buildUnsignedOrder(
	tradingPair string,
	price decimal.Decimal,
	amount decimal.Decimal,
	side string,
	orderType string,
	isMakerOnly bool,
	expireTimeInSecond int64) (orderId string, err error) {
	var dataContainer IBuildOrder
	var body = struct {
		MarketId    string          `json:"marketId"`
		Side        string          `json:"side"`
		OrderType   string          `json:"orderType"`
		Price       decimal.Decimal `json:"price"`
		Amount      decimal.Decimal `json:"amount"`
		Expires     int64           `json:"expires"`
		IsMakerOnly bool            `json:"isMakerOnly"`
		WalletType  string          `json:"walletType"`
	}{
		tradingPair,
		side,
		orderType,
		price,
		amount,
		expireTimeInSecond,
		isMakerOnly,
		"trading",
	}
	bodyBytes, _ := json.Marshal(body)
	resp, err := client.post("orders/build", string(bodyBytes), utils.EmptyKeyPairList)
	if err != nil {
		return "", err
	}
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		err = errors.New(dataContainer.Desc)
	} else {
		orderId = dataContainer.Data.Order.ID
	}

	return
}

func (client *DdexClient) placeOrder(orderId string) (res *OrderRes, err error) {
	var body = struct {
		OrderId   string `json:"orderId"`
		Signature string `json:"signature"`
	}{orderId, client.signOrderId(orderId)}
	bodyBytes, _ := json.Marshal(body)
	resp, err := client.post("orders", string(bodyBytes), utils.EmptyKeyPairList)
	if err != nil {
		return
	}
	var dataContainer IPlaceOrderSync
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		err = errors.New(dataContainer.Desc)
	} else {
		res = client.parseDdexOrderResp(dataContainer.Data.Order)
	}

	return
}

func (client *DdexClient) CreateLimitOrder(
	tradingPair string,
	price decimal.Decimal,
	amount decimal.Decimal,
	side string,
	isMakerOnly bool,
	expireTimeInSecond int64) (orderId string, err error) {
	validPrice := utils.SetDecimal(utils.SetPrecision(price, client.Markets[tradingPair].PricePrecision), client.Markets[tradingPair].PriceDecimal)
	validAmount := utils.SetDecimal(amount, client.Markets[tradingPair].AmountDecimal)
	orderId, err = client.buildUnsignedOrder(tradingPair, validPrice, validAmount, side, "limit", isMakerOnly, expireTimeInSecond)
	if err != nil {
		return
	}
	_, err = client.placeOrder(orderId)
	if err == nil {
		logrus.Debugf("create limit order at %s - price:%s amount:%s side:%s %s", tradingPair, validPrice, validAmount, side, orderId)
	}
	return
}

func (client *DdexClient) CreateMarketOrder(
	tradingPair string,
	priceLimit decimal.Decimal,
	amount decimal.Decimal,
	side string,
) (res *OrderRes, err error) {
	validPrice := utils.SetDecimal(utils.SetPrecision(priceLimit, client.Markets[tradingPair].PricePrecision), client.Markets[tradingPair].PriceDecimal)
	amount = utils.SetDecimal(amount, client.Markets[tradingPair].AmountDecimal)
	orderId, err := client.buildUnsignedOrder(tradingPair, validPrice, amount, side, "market", false, 3600)
	if err != nil {
		return
	}

	res, err = client.placeOrder(orderId)
	if err == nil {
		logrus.Debugf("create market order at %s - price:%s amount:%s side:%s %s", tradingPair, validPrice, amount, side, orderId)
	}
	return
}

func (client *DdexClient) MarketSellAsset(
	tradingPair string,
	assetSymbol string,
	amount decimal.Decimal,
	maxSlippage decimal.Decimal,
) (
	orderId string,
	sellAmount decimal.Decimal,
	receiveAmount decimal.Decimal,
	err error,
) {
	orderId = "0x0"
	sellAmount = decimal.Zero
	receiveAmount = decimal.Zero
	_, _, midPrice, err := client.GetMarketPrice(tradingPair)
	if err != nil {
		return
	}
	var orderRes *OrderRes
	if assetSymbol == strings.Split(tradingPair, "-")[0] {
		orderRes, err = client.CreateMarketOrder(
			tradingPair,
			midPrice.Mul(decimal.New(1, 0).Sub(maxSlippage)),
			amount,
			utils.SELL,
		)
		if err != nil {
			return
		} else {
			orderId = orderRes.Id
			sellAmount = orderRes.FilledAmount
			receiveAmount = orderRes.FilledAmount.Mul(orderRes.AvgPrice)
		}
	} else {
		orderRes, err = client.CreateMarketOrder(
			tradingPair,
			midPrice.Mul(decimal.New(1, 0).Add(maxSlippage)),
			amount,
			utils.BUY,
		)
		if err != nil {
			return
		} else {
			orderId = orderRes.Id
			sellAmount = orderRes.FilledAmount
			receiveAmount = orderRes.FilledAmount.Div(orderRes.AvgPrice)
		}
	}

	return
}

func (client *DdexClient) PromisedMarketSellAsset(
	tradingPair string,
	assetSymbol string,
	amount decimal.Decimal,
	maxSlippage decimal.Decimal,
) (
	orderId string,
	sellAmount decimal.Decimal,
	receiveAmount decimal.Decimal,
	err error,
) {
	for {
		orderId, sellAmount, receiveAmount, err = client.MarketSellAsset(tradingPair, assetSymbol, amount, maxSlippage)
		if err != nil {
			time.Sleep(time.Second)
		} else {
			return
		}
	}
}

func (client *DdexClient) CancelOrder(orderId string) error {
	resp, err := client.delete("orders/"+orderId, utils.EmptyKeyPairList)
	if err != nil {
		return err
	}
	var dataContainer ICancelOrder
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return errors.New(dataContainer.Desc)
	} else {
		logrus.Infof("cancel order %s ", orderId)
		return nil
	}
}

func (client *DdexClient) CancelAllPendingOrders() error {
	for tradingPair, _ := range client.Markets {
		resp, err := client.delete("orders", []utils.KeyPair{{"marketId", tradingPair}})
		if err != nil {
			return err
		}
		var dataContainer ICancelOrder
		json.Unmarshal([]byte(resp), &dataContainer)
		if dataContainer.Desc != "success" {
			return errors.New(dataContainer.Desc)
		}
	}
	logrus.Infof("cancel all orders")
	return nil
}

func (client *DdexClient) parseDdexOrderResp(orderInfo IOrderResp) *OrderRes {
	var orderData = &OrderRes{}
	orderData.Id = orderInfo.ID
	orderData.Amount, _ = decimal.NewFromString(orderInfo.Amount)
	orderData.AvailableAmount, _ = decimal.NewFromString(orderInfo.AvailableAmount)
	orderData.Price, _ = decimal.NewFromString(orderInfo.Price)
	orderData.AvgPrice, _ = decimal.NewFromString(orderInfo.AveragePrice)
	pendingAmount, _ := decimal.NewFromString(orderInfo.PendingAmount)
	confirmedAmount, _ := decimal.NewFromString(orderInfo.ConfirmedAmount)
	orderData.FilledAmount = pendingAmount.Add(confirmedAmount)
	if orderData.AvailableAmount.IsZero() {
		orderData.Status = utils.ORDERCLOSE
	} else {
		orderData.Status = utils.ORDEROPEN
	}
	if orderInfo.Side == "sell" {
		orderData.Side = utils.SELL
	} else {
		orderData.Side = utils.BUY
	}

	return orderData
}

func (client *DdexClient) GetOrder(orderId string) (res *OrderRes, err error) {
	resp, err := client.get("orders/"+orderId, utils.EmptyKeyPairList)
	if err != nil {
		return
	}
	var dataContainer IOrder
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		err = errors.New(dataContainer.Desc)
	} else {
		res = client.parseDdexOrderResp(dataContainer.Data.Order)
	}

	return
}

func (client *DdexClient) GetInventory() (inventory Inventory, err error) {
	inventory = map[string]*Balance{}
	for symbol, asset := range client.Assets {
		var amountHex string
		amountHex, err = client.hydroContract.Call("balanceOf", common.HexToAddress(asset.Address), common.HexToAddress(client.Address))
		if err != nil {
			return
		}
		amount := utils.HexString2Decimal(amountHex, -1*asset.Decimal)
		inventory[symbol] = &Balance{amount, decimal.Zero, amount}
	}

	resp, err := client.get("account/lockedBalances", utils.EmptyKeyPairList)
	if err != nil {
		return
	}
	var dataContainer ILockedBalance
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		err = errors.New(dataContainer.Desc)
		return
	}
	for _, lockedBalance := range dataContainer.Data.LockedBalances {
		if _, ok := inventory[lockedBalance.Symbol]; ok && lockedBalance.WalletType == "trading" {
			amount, _ := decimal.NewFromString(lockedBalance.Amount)
			inventory[lockedBalance.Symbol].Lock = amount.Mul(decimal.New(1, -1*client.Assets[lockedBalance.Symbol].Decimal))
			inventory[lockedBalance.Symbol].Free = inventory[lockedBalance.Symbol].Total.Sub(inventory[lockedBalance.Symbol].Lock)
		}
	}

	return
}

func (client *DdexClient) GetMarketPrice(tradingPair string) (
	bestBidPrice decimal.Decimal,
	bestAskPrice decimal.Decimal,
	midPrice decimal.Decimal,
	err error) {
	resp, err := client.get(
		fmt.Sprintf("markets/%s/orderbook", tradingPair),
		[]utils.KeyPair{{"level", "1"}},
	)
	if err != nil {
		return
	}
	var dataContainer IOrderbook
	json.Unmarshal([]byte(resp), &dataContainer)
	if len(dataContainer.Data.OrderBook.Asks) == 0 || len(dataContainer.Data.OrderBook.Bids) == 0 {
		err = utils.OrderbookNotComplete
		return
	}
	bestAskPrice, _ = decimal.NewFromString(dataContainer.Data.OrderBook.Asks[0].Price)
	bestBidPrice, _ = decimal.NewFromString(dataContainer.Data.OrderBook.Bids[0].Price)
	midPrice = bestAskPrice.Add(bestBidPrice).Div(decimal.New(2, 0))

	return
}

func (client *DdexClient) QuerySellAssetReceiveAmount(
	tradingPair string,
	assetSymbol string,
	payAmount decimal.Decimal,
) (receiveAmount decimal.Decimal, err error) {
	resp, err := client.get(fmt.Sprintf("markets/%s/orderbook", tradingPair), []utils.KeyPair{{"level", "2"}})
	if err != nil {
		return
	}
	var dataContainer IOrderbook
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		err = errors.New(dataContainer.Desc)
		return
	}

	receiveAmount = decimal.Zero
	if assetSymbol == client.Markets[tradingPair].Quote.Symbol {
		for _, ask := range dataContainer.Data.OrderBook.Asks {
			price, _ := decimal.NewFromString(ask.Price)
			amount, _ := decimal.NewFromString(ask.Amount)
			if price.Mul(amount).GreaterThanOrEqual(payAmount) {
				receiveAmount = receiveAmount.Add(payAmount.Div(price))
				payAmount = decimal.Zero
				break
			} else {
				receiveAmount = receiveAmount.Add(amount)
				payAmount = payAmount.Sub(price.Mul(amount))
			}
		}
		if payAmount.IsPositive() {
			err = utils.OrderbookDepthNotEnough
		}
	} else {
		for _, bid := range dataContainer.Data.OrderBook.Bids {
			price, _ := decimal.NewFromString(bid.Price)
			amount, _ := decimal.NewFromString(bid.Amount)
			if amount.GreaterThanOrEqual(payAmount) {
				receiveAmount = receiveAmount.Add(payAmount.Mul(price))
				payAmount = decimal.Zero
				break
			} else {
				receiveAmount = receiveAmount.Add(amount.Mul(price))
				payAmount = payAmount.Sub(amount)
			}
		}
		if payAmount.IsPositive() {
			err = utils.OrderbookDepthNotEnough
		}
	}

	return
}

func (client *DdexClient) GetAssetUSDPrice(assetSymbol string) (price decimal.Decimal, err error) {
	price = decimal.New(-1, 0)
	resp, err := client.get("assets", utils.EmptyKeyPairList)
	if err != nil {
		return
	}
	var dataContainer IAssets
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		err = errors.New(dataContainer.Desc)
		return
	}
	for _, asset := range dataContainer.Data.Assets {
		if assetSymbol == asset.Symbol {
			price, _ = decimal.NewFromString(asset.OracleUSDPrice)
		}
	}
	if price.LessThanOrEqual(decimal.Zero) {
		err = errors.New("not find asset usd price")
	}
	return
}
