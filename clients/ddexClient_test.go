package clients

import (
	"github.com/davecgh/go-spew/spew"
	"os"
	"testing"
)

func TestDdexClient_GetAllPendingOrders(t *testing.T) {
	os.Setenv("ETHEREUM_NODE_URL", "https://ropsten.infura.io/v3/d4470e7b7221494caaaa66d3a353c5dc")
	os.Setenv("CHAIN_ID","3")
	os.Setenv("DDEX_URL", "https://bfd-ropsten-4a7838b8-api.i.ddex.io/v1/")
	os.Setenv("HYDRO_CONTRACT_ADDRESS", "0x06898143DF04616a8A8F9614deb3B99Ba12b3096")
	client,err:=NewDdexClient("B7A0C9D2786FC4DD080EA5D619D36771AEB0C8C26C290AFD3451B92BA2B7BC2C", "ETH","USDT6")
	if err!=nil{
		spew.Dump(err)
	}
	bid, ask, mid ,err:=client.GetMarketPrice()
	spew.Dump(bid,ask,mid,err)
}