package clients

import (
	"github.com/davecgh/go-spew/spew"
	"os"
	"testing"
)

func TestBidderClient_GetSingleAuction(t *testing.T) {
	os.Setenv("ETHEREUM_NODE_URL", "https://mainnet.infura.io/v3/d4470e7b7221494caaaa66d3a353c5dc")
	os.Setenv("CHAIN_ID","1")
	os.Setenv("HYDRO_CONTRACT_ADDRESS", "0x241e82C79452F51fbfc89Fac6d912e021dB1a3B7")

	client,err:=NewBidderClient()
	if err!=nil{
		spew.Dump(err)
	}

	spew.Dump(client.GetSingleAuction(2))

}