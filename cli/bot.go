package cli

import (
	"auctionBidder/client"
	"auctionBidder/utils"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"strings"
)

type BidderBot struct {
	BidderClient     *client.BidderClient
	DdexClient       *client.DdexClient
	BlockChannel     chan int64
	MaxSlippage      decimal.Decimal
	MinOrderValueUSD decimal.Decimal
	GasPriceTipsGwei int // send tx using gas price "fast" from ether gas station plus tips
	Markets          string
	ProfitBuffer     decimal.Decimal
}

func (b *BidderBot) Run() {
	b.updatePnlView()
	for true {
		blockNum := <-b.BlockChannel
		logrus.Infof("new block %d", blockNum)
		allAuctions, err := b.BidderClient.GetAllAuctions()
		if err != nil {
			continue
		}
		UpdateAuctionView(allAuctions)
		for _, auction := range allAuctions {
			err := b.tryFillAuction(auction)
			if err != nil {
				logrus.Warnf("try fill auction #%d failed: %s", auction.ID, err.Error())
			}
		}
	}
}

func (b *BidderBot) tryFillAuction(auction *client.Auction) (err error) {
	// check if the market is under monitor
	if !strings.Contains(b.Markets, auction.TradingPair) {
		return nil
	}
	logrus.Debugf("try fill auction %d", auction.ID)

	// truncate order size by free balance
	inventory, err := b.DdexClient.GetInventory()
	if err != nil {
		return
	}
	freeBalance := inventory[auction.DebtSymbol].Free
	if freeBalance.IsZero() {
		err = errors.Errorf(`%s balance is zero`, auction.DebtSymbol)
		return
	}

	var debt = auction.AvailableDebt
	var collateral = auction.AvailableCollateral
	if freeBalance.LessThan(debt) {
		logrus.Warnf(`Balance not enough, you could repay %s %s debt`, freeBalance.String(), auction.DebtSymbol)
		debt = freeBalance
		collateral = debt.Div(auction.AvailableDebt).Mul(auction.AvailableCollateral)
	}

	// amount must greater than min usd size
	collateralPrice, err := b.DdexClient.GetAssetUSDPrice(auction.CollateralSymbol)
	collateralValue := collateral.Mul(collateralPrice)
	if err == nil && collateral.Mul(collateralPrice).LessThanOrEqual(b.MinOrderValueUSD) {
		err = errors.Errorf("collateral usd value %s$ too small", collateralValue.String())
		return
	}

	// check auction profitable
	receive, err := b.DdexClient.QuerySellAssetReceiveAmount(auction.TradingPair, auction.CollateralSymbol, collateral)
	if err != nil {
		return
	}

	if receive.LessThanOrEqual(debt.Add(debt.Mul(b.ProfitBuffer))) {
		err = errors.Errorf("auction price not profitable, wait next block")
		return
	} else {
		logrus.Infof("auction price profitable!")
		gasPriceInGwei := b.getGasPrice()
		txHash, err := b.BidderClient.FillAuctioon(auction, debt, gasPriceInGwei)
		if err != nil {
			return err
		}
		logrus.Infof("send tx %s", txHash)

		bidderRepay, collateralForBidder, gasUsed, err := b.BidderClient.GetFillAuctionRes(txHash, auction)
		gasCost := gasUsed.Mul(decimal.New(gasPriceInGwei, -9))
		logrus.Infof(
			"fill auction: repayDebt %s%s receiveCollateral %s%s gasCost %sETH",
			bidderRepay.String(),
			auction.DebtSymbol,
			collateralForBidder.String(),
			auction.CollateralSymbol,
			gasCost.String())

		if collateralForBidder.IsZero() {
			utils.InsertFailedBid(txHash, int(auction.ID), auction.DebtSymbol, auction.CollateralSymbol, gasCost.String())
			b.updatePnlView()
			err = errors.New("bid transaction failed")
			return err
		}
		ddexOrderId, ddexSellCollateral, ddexReceiveDebt, err := b.DdexClient.PromisedMarketSellAsset(auction.TradingPair, auction.CollateralSymbol, collateralForBidder, b.MaxSlippage)
		logrus.Infof("hedge at ddex market: sell %s%s receive %s%s",
			ddexSellCollateral.String(),
			auction.CollateralSymbol,
			ddexReceiveDebt.String(),
			auction.DebtSymbol,
		)

		if err != nil {
			return err
		}
		utils.InsertAuctionRes(
			txHash,
			int(auction.ID),
			auction.DebtSymbol,
			auction.CollateralSymbol,
			bidderRepay.String(),
			collateralForBidder.String(),
			ddexOrderId,
			ddexSellCollateral.String(),
			ddexReceiveDebt.String(),
			gasCost.String(),
		)
		b.updatePnlView()
	}

	return
}

func (b *BidderBot) getGasPrice() int64 {
	resp, err := utils.Get("https://ethgasstation.info/json/ethgasAPI.json", "", utils.EmptyKeyPairList, utils.EmptyKeyPairList)
	if err != nil {
		return 40000000000 // default 40gwei
	}
	var dataContainer struct {
		Fast    int `json:"fast"`
		Fastest int `json:"fastest"`
		SafeLow int `json:"safeLow"`
		Average int `json:"average"`
	}
	json.Unmarshal([]byte(resp), &dataContainer)
	return int64(dataContainer.Fast/10 + b.GasPriceTipsGwei)
}

func (b *BidderBot) updatePnlView() {
	position, err := utils.QueryPosition()
	if err != nil {
		return
	}
	var price = map[string]decimal.Decimal{}
	for symbol, _ := range position {
		p, err := b.DdexClient.GetAssetUSDPrice(symbol)
		if err != nil {
			p = decimal.Zero
		}
		price[symbol] = p
	}
	UpdatePnlView(position, price)
}
