package clients

import (
	"auctionBidder/utils"
	"auctionBidder/web3"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
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
	Amount          decimal.Decimal
	Price           decimal.Decimal
	AvailableAmount decimal.Decimal
	FilledAmount    decimal.Decimal
	AvgPrice        decimal.Decimal
	Side            string
}

type Balance struct {
	Free  decimal.Decimal
	Lock  decimal.Decimal
	Total decimal.Decimal
}

type Inventory struct {
	Quote Balance
	Base  Balance
}

type DdexClient struct {
	hydroContract     *web3.Contract
	privateKey        string
	signCache         string
	lastSignTime      int64
	address           string
	tradingPair       string
	quoteTokenSymbol  string
	quoteTokenAddress string
	quoteTokenDecimal int
	baseTokenSymbol   string
	baseTokenAddress  string
	baseTokenDecimal  int
	baseUrl           string
	pricePrecision    int
	priceDecimal      int
	amountDecimal     int
	minAmount         decimal.Decimal
}

func NewDdexClient(privateKey string, baseTokenSymbol string, quoteTokenSymbol string) (client *DdexClient, err error) {
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
	var dataContainer IDdexMarkets
	resp, err := utils.Get(
		utils.JoinUrlPath(ddexBaseUrl, fmt.Sprintf("markets/%s-%s", baseTokenSymbol, quoteTokenSymbol)),
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
	minAmount, _ := decimal.NewFromString(dataContainer.Data.Market.MinOrderSize)

	client = &DdexClient{
		contract,
		privateKey,
		"",
		0,
		address,
		fmt.Sprintf("%s-%s", baseTokenSymbol, quoteTokenSymbol),
		quoteTokenSymbol,
		dataContainer.Data.Market.QuoteAssetAddress,
		dataContainer.Data.Market.QuoteAssetDecimals,
		baseTokenSymbol,
		dataContainer.Data.Market.BaseAssetAddress,
		dataContainer.Data.Market.BaseAssetDecimals,
		ddexBaseUrl,
		dataContainer.Data.Market.PricePrecision,
		dataContainer.Data.Market.PriceDecimals,
		dataContainer.Data.Market.AmountDecimals,
		minAmount,
	}

	return
}

func (client *DdexClient) updateSignCache() {
	now := utils.MillisecondTimestamp()
	if client.lastSignTime < now-200000 {
		messageStr := "HYDRO-AUTHENTICATION@" + strconv.Itoa(int(now))
		signRes, _ := utils.PersonalSign([]byte(messageStr), client.privateKey)
		client.signCache = fmt.Sprintf("%s#%s#0x%x", strings.ToLower(client.address), messageStr, signRes)
		client.lastSignTime = now
	}
}

func (client *DdexClient) signOrderId(orderId string) string {
	if strings.HasPrefix(orderId, "0x") {
		orderId = orderId[2:]
	}
	orderIdHex, _ := hex.DecodeString(orderId)
	signature, _ := utils.PersonalSign(orderIdHex, client.privateKey)
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
	price decimal.Decimal,
	amount decimal.Decimal,
	side string,
	orderType string,
	isMakerOnly bool,
	expireTimeInSecond int64) (string, error) {
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
	}{client.tradingPair, side, orderType, price, amount, expireTimeInSecond, isMakerOnly, "trading"}
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
	var body = struct {
		OrderId   string `json:"orderId"`
		Signature string `json:"signature"`
	}{orderId, client.signOrderId(orderId)}
	bodyBytes, _ := json.Marshal(body)
	resp, err := client.post("orders", string(bodyBytes), utils.EmptyKeyPairList)
	if err != nil {
		return false
	}
	var dataContainer IPlaceOrder
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		spew.Dump(dataContainer)
		return false
	} else {
		return true
	}
}

func (client *DdexClient) placeOrderSynchronously(orderId string) (res *OrderRes, err error) {
	var body = struct {
		OrderId   string `json:"orderId"`
		Signature string `json:"signature"`
	}{orderId, client.signOrderId(orderId)}
	bodyBytes, _ := json.Marshal(body)
	resp, err := client.post("orders/sync", string(bodyBytes), utils.EmptyKeyPairList)
	if err != nil {
		return
	}
	var dataContainer IPlaceOrderSync
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		err = errors.New(dataContainer.Desc)
		return
	} else {
		res = client.parseDdexOrderResp(dataContainer.Data.Order)
		return
	}
}

func (client *DdexClient) CreateLimitOrder(
	price decimal.Decimal,
	amount decimal.Decimal,
	side string,
	isMakerOnly bool,
	expireTimeInSecond int64) (string, error) {

	validPrice := utils.SetDecimal(utils.SetPrecision(price, client.pricePrecision), client.priceDecimal)

	validAmount := utils.SetDecimal(amount, client.amountDecimal)
	if validAmount.Mul(validPrice).LessThan(client.minAmount) {
		return "", errors.New(fmt.Sprintf("ddex client %s create order amount %s less than min amount %s", client.tradingPair, validAmount.String(), client.minAmount.String()))
	}

	orderId, err := client.buildUnsignedOrder(validPrice, validAmount, side, "limit", isMakerOnly, expireTimeInSecond)
	if err != nil {
		return "", err
	}
	placeSuccess := client.placeOrder(orderId)
	if placeSuccess {
		logrus.Infof("ddex client %s create limit order - price:%s amount:%s side:%s %s", client.tradingPair, validPrice, validAmount, side, orderId)
		return orderId, nil
	} else {
		return "", errors.New(fmt.Sprintf("ddex client %s place order failed", client.tradingPair))
	}
}

func (client *DdexClient) CreateMarketOrder(
	priceLimit decimal.Decimal,
	amount decimal.Decimal,
	side string,
) (res *OrderRes, err error) {

	validPrice := utils.SetDecimal(utils.SetPrecision(priceLimit, client.pricePrecision), client.priceDecimal)
	if side == utils.SELL {
		amount = utils.SetDecimal(amount, client.amountDecimal)
	}

	orderId, err := client.buildUnsignedOrder(validPrice, amount, side, "market", false, 3600)
	if err != nil {
		return
	}
	res, err = client.placeOrderSynchronously(orderId)
	return
}

func (client *DdexClient) CancelOrder(orderId string) error {
	resp, err := client.delete("orders/"+orderId, utils.EmptyKeyPairList)
	if err != nil {
		return err
	}
	var dataContainer ICancelOrder
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return errors.New(fmt.Sprintf("ddex client %s cancel order %s failed", client.tradingPair, orderId))
	} else {
		logrus.Infof("ddex client %s cancel order %s succeed", client.tradingPair, orderId)
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
	resp, err := client.delete("orders", []utils.KeyPair{{"marketId", client.tradingPair}})
	if err != nil {
		return false, err
	}
	var dataContainer ICancelOrder
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		return false, errors.New(fmt.Sprintf("ddex client %s cancel all pending orders failed", client.tradingPair))
	} else {
		logrus.Infof("ddex client %s cancel all orders succeed", client.tradingPair)
		return true, nil
	}
}

func (client *DdexClient) parseDdexOrderResp(orderInfo IDDEXOrderResp) *OrderRes {
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

func (client *DdexClient) GetOrder(orderId string) (res *OrderRes, err error) {
	resp, err := client.get("orders/"+orderId, utils.EmptyKeyPairList)
	if err != nil {
		return
	}
	var dataContainer IOrder
	json.Unmarshal([]byte(resp), &dataContainer)
	if dataContainer.Desc != "success" {
		err = errors.New(dataContainer.Desc)
		return
	} else {
		return client.parseDdexOrderResp(dataContainer.Data.Order), nil
	}
}

func (client *DdexClient) PromisedGetClosedOrder(orderId string) *OrderRes {
	for true {
		info, err := client.GetOrder(orderId)
		if err == nil && info.Status == utils.ORDER_CLOSE {
			return info
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (client *DdexClient) GetAllPendingOrders() ([]*OrderRes, error) {
	var allOrders = []*OrderRes{}
	var pageNum = 1
	for true {
		resp, err := client.get("orders", []utils.KeyPair{
			{"marketId", client.tradingPair},
			{"perPage", "100"},
			{"page", strconv.Itoa(pageNum)},
		})
		if err != nil {
			return allOrders, err
		}
		var dataContainer IAllPendingOrders
		json.Unmarshal([]byte(resp), &dataContainer)
		if dataContainer.Desc != "success" {
			return allOrders, errors.New(fmt.Sprintf("ddex client %s get all pending orders failed", client.tradingPair))
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

func (client *DdexClient) GetInventory() (inventory *Inventory, err error) {
	baseAmountHex, err := client.hydroContract.Call("balanceOf", common.HexToAddress(client.baseTokenAddress), common.HexToAddress(client.address))
	if err != nil {
		return
	}
	quoteAmountHex, err := client.hydroContract.Call("balanceOf", common.HexToAddress(client.quoteTokenAddress), common.HexToAddress(client.address))
	if err != nil {
		return
	}
	baseAmount := decimal.NewFromBigInt(utils.Hex2BigInt(baseAmountHex), int32(-1*client.baseTokenDecimal))
	quoteAmount := decimal.NewFromBigInt(utils.Hex2BigInt(quoteAmountHex), int32(-1*client.quoteTokenDecimal))

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
	lockedBase := decimal.Zero
	lockedQuote := decimal.Zero
	for _, lockedBalance := range dataContainer.Data.LockedBalances {
		if lockedBalance.AssetAddress == client.baseTokenAddress && lockedBalance.WalletType == "trading" {
			amount, _ := decimal.NewFromString(lockedBalance.Amount)
			lockedBase = amount.Mul(decimal.New(1, int32(-1*client.baseTokenDecimal)))
		}
		if lockedBalance.AssetAddress == client.quoteTokenAddress && lockedBalance.WalletType == "trading" {
			amount, _ := decimal.NewFromString(lockedBalance.Amount)
			lockedQuote = amount.Mul(decimal.New(1, int32(-1*client.quoteTokenDecimal)))
		}
	}

	inventory = &Inventory{
		Balance{
			baseAmount.Sub(lockedBase),
			lockedBase,
			baseAmount,
		},
		Balance{
			quoteAmount.Sub(lockedQuote),
			lockedQuote,
			quoteAmount,
		},
	}
	return
}

func (client *DdexClient) GetMarketPrice() (
	bestBidPrice decimal.Decimal,
	bestAskPrice decimal.Decimal,
	midPrice decimal.Decimal,
	err error) {
	resp, err := client.get(
		fmt.Sprintf("markets/%s/orderbook", client.tradingPair),
		[]utils.KeyPair{{"level", "1"}},
	)
	if err != nil {
		return
	}
	var dataContainer ILevel1Orderbook
	json.Unmarshal([]byte(resp), &dataContainer)
	if len(dataContainer.Data.OrderBook.Asks) == 0 || len(dataContainer.Data.OrderBook.Bids) == 0 {
		err = errors.New("current orderbook not complete")
		return
	}
	bestAskPrice, _ = decimal.NewFromString(dataContainer.Data.OrderBook.Asks[0].Price)
	bestBidPrice, _ = decimal.NewFromString(dataContainer.Data.OrderBook.Bids[0].Price)
	midPrice = bestAskPrice.Add(bestBidPrice).Div(decimal.New(2, 0))

	return
}
