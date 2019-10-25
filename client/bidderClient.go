package client

import (
	"auctionBidder/utils"
	"auctionBidder/web3"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"os"
	"strings"
	"time"
)

type Auction struct {
	ID                  int64
	DebtSymbol          string
	CollateralSymbol    string
	TradingPair         string
	AvailableDebt       decimal.Decimal
	AvailableCollateral decimal.Decimal
	Ratio               decimal.Decimal
	Price               decimal.Decimal
	Finished            bool
}

type BidderClient struct {
	web3             *web3.Web3
	hydroContract    *web3.Contract
	bidderPrivateKey string
	bidderAddress    string
	assets           map[string]*Asset  // symbol -> asset
	markets          map[string]*Market // trading pair -> market
}

func NewBidderClient(bidderPrivateKey string, assets map[string]*Asset, markets map[string]*Market) (client *BidderClient, err error) {
	ethereumNodeUrl := os.Getenv("ETHEREUM_NODE_URL")
	hydroContractAddress := os.Getenv("HYDRO_CONTRACT_ADDRESS")

	web3 := web3.NewWeb3(ethereumNodeUrl)
	bidderAddress, err := web3.AddPrivateKey(bidderPrivateKey)
	if err != nil {
		return
	}
	contract, err := web3.NewContract(utils.HydroAbi, hydroContractAddress)
	if err != nil {
		return
	}
	client = &BidderClient{
		web3,
		contract,
		bidderPrivateKey,
		bidderAddress,
		assets,
		markets,
	}

	return
}

func (client *BidderClient) FillAuction(
	auction *Auction,
	repayDebt decimal.Decimal,
	gasPriceInGwei int64,
) (txHash string, err error) {
	nonce, err := client.web3.Rpc.EthGetTransactionCount(client.bidderAddress, "latest")
	if err != nil {
		return
	}
	sendTxParams := &web3.SendTxParams{
		client.bidderAddress,
		big.NewInt(500000),
		big.NewInt(gasPriceInGwei * 1000000000),
		uint64(nonce),
	}

	rawRepayDebt := repayDebt.Mul(decimal.New(1, client.assets[auction.DebtSymbol].Decimal)).Floor()

	return client.hydroContract.Send(sendTxParams, big.NewInt(0), "fillAuctionWithAmount", uint32(auction.ID), utils.DecimalToBigInt(rawRepayDebt))
}

func (client *BidderClient) GetFillAuctionRes(txHash string, auction *Auction) (
	bidderRepay decimal.Decimal,
	collateralForBidder decimal.Decimal,
	gasUsed decimal.Decimal,
	err error) {
	var receipt *web3.TransactionReceipt
	for true {
		receipt, err = client.web3.Rpc.EthGetTransactionReceipt(txHash)
		if err != nil || receipt.BlockNumber == 0 {
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	gasUsed = decimal.New(int64(receipt.GasUsed), 0)
	if receipt.Status == "0x0" {
		bidderRepay = decimal.Zero
		collateralForBidder = decimal.Zero
		return
	}
	for _, log := range receipt.Logs {
		if log.Topics[0] == "0x42a553656a0da7239e70a4a3c864c1ac7d46d7968bfe2e1fb14f42dbb67135e8" {
			bidderRepay = utils.HexString2Decimal(log.Data[130:194], -1*client.assets[auction.DebtSymbol].Decimal)
			collateralForBidder = utils.HexString2Decimal(log.Data[194:258], -1*client.assets[auction.CollateralSymbol].Decimal)
			return
		}
	}

	return
}

func (client *BidderClient) GetCurrentAuctionIDs() (auctionIDs []int64, err error) {
	resp, err := client.hydroContract.Call("getCurrentAuctions")
	if err != nil {
		return nil, err
	}
	auctionNum := (len(resp)-2)/64 - 2
	for i := 0; i < auctionNum; i++ {
		auctionID, _ := utils.HexString2Int(resp[130+64*i : 194+64*i])
		auctionIDs = append(auctionIDs, int64(auctionID))
	}

	return
}

func (client *BidderClient) GetSingleAuction(auctionID int64) (auction *Auction, err error) {
	resp, err := client.hydroContract.Call("getAuctionDetails", uint32(auctionID))
	if err != nil {
		return
	}

	if len(resp) != 578 {
		err = utils.AuctionNotExist
		return
	}

	debtAddress := "0x" + strings.ToLower(resp[2+64*2+24:2+64*3])
	collateralAddress := "0x" + strings.ToLower(resp[2+64*3+24:2+64*4])

	var debtSymbol string
	var collateralSymbol string

	for symbol, asset := range client.assets {
		if utils.IsAddressEqual(asset.Address, debtAddress) {
			debtSymbol = symbol
		}
		if utils.IsAddressEqual(asset.Address, collateralAddress) {
			collateralSymbol = symbol
		}
	}

	availableCollateral := utils.HexString2Decimal(resp[2+64*5:2+64*6], -1*client.assets[collateralSymbol].Decimal)
	availableDetb := utils.HexString2Decimal(resp[2+64*4:2+64*5], -1*client.assets[debtSymbol].Decimal)
	availableDetb = availableDetb.Mul(decimal.New(1, 0).Add(decimal.New(1, -5))) // the debt is growing while auction ongoing

	ratio := utils.HexString2Decimal(resp[2+64*6:2+64*7], -18)
	finished := utils.HexString2Decimal(resp[2+64*8:2+64*9], 0).IsPositive()

	if ratio.LessThan(decimal.New(1, 0)) {
		availableCollateral = availableCollateral.Mul(ratio)
	} else {
		availableDetb = availableDetb.Div(ratio)
	}

	var price decimal.Decimal
	if availableCollateral.IsPositive() {
		price = availableDetb.Div(availableCollateral)
	} else {
		price = decimal.Zero
	}

	var tradingPair string
	if _, ok := client.markets[fmt.Sprintf("%s-%s", debtSymbol, collateralSymbol)]; ok {
		tradingPair = fmt.Sprintf("%s-%s", debtSymbol, collateralSymbol)
	} else {
		tradingPair = fmt.Sprintf("%s-%s", collateralSymbol, debtSymbol)
	}

	auction = &Auction{
		auctionID,
		debtSymbol,
		collateralSymbol,
		tradingPair,
		availableDetb,
		availableCollateral,
		ratio,
		price,
		finished,
	}

	return
}

func (client *BidderClient) GetAllAuctions() (auctions []*Auction, err error) {
	auctionIDs, err := client.GetCurrentAuctionIDs()
	if err != nil {
		return
	}
	auctions = []*Auction{}
	for _, auctionID := range auctionIDs {
		auction, err := client.GetSingleAuction(auctionID)
		if err == nil {
			auctions = append(auctions, auction)
		}
	}

	return
}
