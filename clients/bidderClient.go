package clients

import (
	"auctionBidder/utils"
	"auctionBidder/web3"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"os"
	"strings"
)


type Auction struct {
	ID              int64
	DebtAsset       string
	CollateralAsset string
	AvailableDebt        decimal.Decimal
	AvailableCollateral  decimal.Decimal
	Ratio           decimal.Decimal
	Price           decimal.Decimal
	Finished        bool
}

type AuctionList []*Auction

func (l AuctionList) Len() int {
	return len(l)
}

func (l AuctionList) Less(i, j int) bool {
	return l[i].Price.LessThan(l[j].Price)
}

func (l AuctionList) Swap(i, j int) {
	temp := l[i]
	l[i] = l[j]
	l[j] = temp
}

type BidderClient struct {
	web3 *web3.Web3
	hydroContract *web3.Contract
}

func NewBidderClient() (client *BidderClient, err error){
	ethereumNodeUrl:=os.Getenv("ETHEREUM_NODE_URL")
	hydroContractAddress:=os.Getenv("HYDRO_CONTRACT_ADDRESS")

	web3:=web3.NewWeb3(ethereumNodeUrl)
	contract,err:=web3.NewContract(utils.HydroAbi, hydroContractAddress)
	if err!=nil {
		return
	}
	client = &BidderClient{
		web3,
		contract,
	}
	return
}

// return raw amount
func (client *BidderClient) GetFillAuctionRes(txHash string) (bidderRepay decimal.Decimal, collateralForBidder decimal.Decimal, err error) {
	receipt, err := client.web3.Rpc.EthGetTransactionReceipt(txHash)
	if err != nil {
		return
	}
	if receipt.Status!="0x1" {
		return decimal.Zero, decimal.Zero, nil
	}
	for _, log := range receipt.Logs {
		if log.Topics[0] == "0x42a553656a0da7239e70a4a3c864c1ac7d46d7968bfe2e1fb14f42dbb67135e8" {
			spew.Dump(log)
			bidderRepay = decimal.NewFromBigInt(utils.Hex2BigInt(log.Data[130:194]), 0)
			collateralForBidder = decimal.NewFromBigInt(utils.Hex2BigInt(log.Data[194:258]), 0)
			return
		}
	}
	err = errors.New("not find fill auction event")
	return
}

func (client *BidderClient) GetCurrentAuctionIDs() (auctionIDs []int64, err error) {
	resp,err:=client.hydroContract.Call("getCurrentAuctions")
	if err != nil {
		return nil, err
	}
	// spew.Dump(dataContainer)
	auctionNum := (len(resp)-2)/64 - 2
	for i := 0; i < auctionNum; i++ {
		auctionIDs = append(auctionIDs, utils.Hex2BigInt(resp[130+64*i:194+64*i]).Int64())
	}
	return auctionIDs, nil
}

func (client *BidderClient) GetSingleAuction(auctionID int64) (auction *Auction, err error) {
	resp,err:=client.hydroContract.Call("getAuctionDetails", uint32(auctionID))
	if err!=nil{
		return
	}

	if len(resp) != 578 {
		err = errors.New("auction details length wrong")
		return
	}

	debtAsset := "0x" + strings.ToLower(resp[2+64*2+24:2+64*3])
	collateralAsset := "0x" + strings.ToLower(resp[2+64*3+24:2+64*4])
	availableDetb := decimal.NewFromBigInt(utils.Hex2BigInt(resp[2+64*4:2+64*5]), 0)
	availableCollateral := decimal.NewFromBigInt(utils.Hex2BigInt(resp[2+64*5:2+64*6]), 0)

	ratio := decimal.NewFromBigInt(utils.Hex2BigInt(resp[2+64*6:2+64*7]), -18)
	finished := utils.Hex2BigInt(resp[2+64*8:2+64*9]).Uint64()==1

	if ratio.LessThan(decimal.New(1,0)){
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

	auction = &Auction{
		auctionID,
		debtAsset,
		collateralAsset,
		availableDetb,
		availableCollateral,
		ratio,
		price,
		finished,
	}

	return auction, nil
}

