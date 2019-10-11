package web3

import (
	"auctionBidder/utils"
	"errors"
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

func (w *Web3) NewContract(abiStr string, address string) (*Contract, error) {
	abi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, err
	}
	commonAddress := common.HexToAddress(address)
	return &Contract{
		w, &abi, &commonAddress,
	}, nil
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
		err = errors.New(fmt.Sprintf("from address %s not exist", params.FromAddress))
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

	chainID := os.Getenv("CHAIN_ID")
	if len(chainID) == 0 {
		panic("need env CHAIN_ID")
	}
	rawData, err := utils.SignTx(c.web3.privateKeyMap[strings.ToLower(params.FromAddress)], os.Getenv("CHAIN_ID"), tx)
	if err != nil {
		panic(err)
	}
	return c.web3.Rpc.EthSendRawTransaction(rawData)
}
