package web3

import (
	"auctionBidder/utils"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"math/big"
	"os"
	"testing"
)

func TestWeb3(t *testing.T) {
	os.Setenv("CHAIN_ID", "3")
	web3 := NewWeb3("https://ropsten.infura.io/v3/d4470e7b7221494caaaa66d3a353c5dc")
	contract, _ := web3.NewContract(utils.Erc20Abi, "0x818375a1de08b5fd7cd0b919dc8d4c30acfb7fe8")
	web3.AddPrivateKey("B7A0C9D2786FC4DD080EA5D619D36771AEB0C8C26C290AFD3451B92BA2B7BC2C")
	nonce, err := web3.Rpc.EthGetTransactionCount("0x31Ebd457b999Bf99759602f5Ece5AA5033CB56B3", "latest")
	resp, err := contract.Send(&SendTxParams{
		"0x31Ebd457b999Bf99759602f5Ece5AA5033CB56B3",
		big.NewInt(100000),
		utils.DecimalToBigInt(decimal.New(20, 9)),
		uint64(nonce),
	},
		big.NewInt(0),
		"transfer",
		common.HexToAddress("0x9c59990ec0177d87ED7D60A56F584E6b06C639a2"),
		utils.DecimalToBigInt(decimal.New(1, 18)),
	)
	spew.Dump(resp, err)
}

func TestGetReceipt(t *testing.T) {
	os.Setenv("CHAIN_ID", "1")
	web3 := NewWeb3("https://mainnet.infura.io/v3/d4470e7b7221494caaaa66d3a353c5dc")
	spew.Dump(web3.Rpc.EthGetTransactionReceipt("0xc36bc99a2ea20504b1fde73ae7d48c0f0779cb727ab85b062e6addc5dfae2fe1"))
}
