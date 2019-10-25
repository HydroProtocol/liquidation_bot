package web3

import (
	"auctionBidder/utils"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"os"
	"strings"
)

type Web3 struct {
	Rpc           *EthRPC
	privateKeyMap map[string]string // address -> privateKey
}

func NewWeb3(ethereumNodeUrl string) *Web3 {
	rpc := NewEthRPC(ethereumNodeUrl)

	return &Web3{rpc, map[string]string{}}
}

func (w *Web3) AddPrivateKey(privateKey string) (newAddress string, err error) {
	pk, err := utils.NewPrivateKeyByHex(privateKey)
	if err != nil {
		return
	}
	newAddress = utils.PubKey2Address(pk.PublicKey)
	w.privateKeyMap[strings.ToLower(newAddress)] = strings.ToLower(privateKey)

	return
}

func (w *Web3) NewBlockChannel() chan int64 {
	c := make(chan int64)
	go func() {
		blockNum := 0
		for true {
			newBlockNum, err := w.Rpc.EthBlockNumber()
			if err == nil {
				if newBlockNum > blockNum {
					c <- int64(newBlockNum)
					blockNum = newBlockNum
				}
			}
		}
	}()

	return c
}

type SendTxParams struct {
	FromAddress string
	GasLimit    *big.Int
	GasPrice    *big.Int
	Nonce       uint64
}

type Contract struct {
	web3    *Web3
	abi     *abi.ABI
	address *common.Address
}

func (w *Web3) NewContract(abiStr string, address string) (contract *Contract, err error) {
	abi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return
	}

	commonAddress := common.HexToAddress(address)
	contract = &Contract{
		w, &abi, &commonAddress,
	}

	return
}

func (c *Contract) Call(functionName string, args ...interface{}) (resp string, err error) {
	var dataByte []byte
	if args != nil {
		dataByte, err = c.abi.Pack(functionName, args...)
	} else {
		dataByte = c.abi.Methods[functionName].ID()
	}
	if err != nil {
		return
	}

	return c.web3.Rpc.EthCall(T{
		To:   c.address.String(),
		From: "0x0000000000000000000000000000000000000000",
		Data: fmt.Sprintf("0x%x", dataByte)},
		"latest",
	)
}

func (c *Contract) Send(params *SendTxParams, amount *big.Int, functionName string, args ...interface{}) (resp string, err error) {
	if _, ok := c.web3.privateKeyMap[strings.ToLower(params.FromAddress)]; !ok {
		err = utils.AddressNotExist
		return
	}

	data, err := c.abi.Pack(functionName, args...)
	if err != nil {
		return
	}

	tx := types.NewTransaction(
		params.Nonce,
		*c.address,
		amount,
		params.GasLimit.Uint64(),
		params.GasPrice,
		data,
	)
	rawData, _ := utils.SignTx(c.web3.privateKeyMap[strings.ToLower(params.FromAddress)], os.Getenv("CHAIN_ID"), tx)

	return c.web3.Rpc.EthSendRawTransaction(rawData)
}

func GetGasPriceGwei() (gasPriceInGwei int64) {
	resp, err := utils.Get("https://ethgasstation.info/json/ethgasAPI.json", "", utils.EmptyKeyPairList, utils.EmptyKeyPairList)
	if err != nil {
		return 30 // default 30gwei
	}
	var dataContainer struct {
		Fast    float64 `json:"fast"`
		Fastest float64 `json:"fastest"`
		SafeLow float64 `json:"safeLow"`
		Average float64 `json:"average"`
	}
	json.Unmarshal([]byte(resp), &dataContainer)
	gasPriceInGwei = int64(dataContainer.Fast / 10)
	if gasPriceInGwei > 300 {
		gasPriceInGwei = 300
	}
	return
}
