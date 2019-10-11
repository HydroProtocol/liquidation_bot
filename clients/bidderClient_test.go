package clients

import (
	"github.com/davecgh/go-spew/spew"
	"os"
	"testing"
)

func TestBidderClient_GetSingleAuction(t *testing.T) {
	os.Setenv("ETHEREUM_NODE_URL", "https://ropsten.infura.io/v3/d4470e7b7221494caaaa66d3a353c5dc")
	os.Setenv("CHAIN_ID","3")
	os.Setenv("HYDRO_CONTRACT_ADDRESS", "0x06898143DF04616a8A8F9614deb3B99Ba12b3096")

	client,err:=NewBidderClient("B7A0C9D2786FC4DD080EA5D619D36771AEB0C8C26C290AFD3451B92BA2B7BC2C")
	if err!=nil{
		spew.Dump(err)
	}

	// spew.Dump(client.FillAuctioon(1, decimal.New(1,0), 20000000000))
	spew.Dump(client.GetFillAuctionRes("0x2f61f18ddfcb26a72c438f9e33597b52794f835ba2320b6f674381bf7dfbb934"))
}