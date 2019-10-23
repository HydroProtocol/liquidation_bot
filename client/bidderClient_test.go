package client

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
	"os"
	"testing"
)

func Test_temp(t *testing.T) {
	os.Setenv("ETHEREUM_NODE_URL", "https://ropsten.infura.io/v3/d4470e7b7221494caaaa66d3a353c5dc")
	os.Setenv("CHAIN_ID", "3")
	os.Setenv("DDEX_URL", "https://bfd-ropsten-59c1702d-api.intra.ddex.io/v4/")
	os.Setenv("HYDRO_CONTRACT_ADDRESS", "0x06898143DF04616a8A8F9614deb3B99Ba12b3096")
	pk := "B7A0C9D2786FC4DD080EA5D619D36771AEB0C8C26C290AFD3451B92BA2B7BC2C"
	client, err := NewDdexClient(pk)
	if err != nil {
		spew.Dump(err)
	}

	spew.Dump(client.MarketSellAsset("ETH-USDT6", "USDT6", decimal.New(100, 0), decimal.New(5, -1)))

	// bidder, err := NewBidderClient(pk, client.Assets)
	// if err != nil {
	// 	spew.Dump(err)
	// }
	// spew.Dump(bidder.GetAllAuctions())
}
