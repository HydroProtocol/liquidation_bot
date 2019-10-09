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
	rpc *EthRPC
	privateKeyMap map[string]string // address -> privateKey
}

func NewWeb3(ethereumNodeUrl string) *Web3 {
	rpc:=NewEthRPC(ethereumNodeUrl)
	return &Web3{rpc,map[string]string{}}
}

func (w *Web3)AddPrivateKey(privateKey string) error {
	pk,err:=utils.NewPrivateKeyByHex(privateKey)
	if err!=nil {
		return err
	}
	address:=utils.PubKey2Address(pk.PublicKey)
	w.privateKeyMap[strings.ToLower(address)] = strings.ToLower(privateKey)
	return nil
}

type SendTxParams struct {
	fromAddress string
	gasLimit *big.Int
	gasPrice *big.Int
	nonce uint64
}

type Contract struct {
	web3 *Web3
	abi    *abi.ABI
	address *common.Address
}

func (w *Web3)NewContract(abiStr string, address string) (*Contract,error) {
	abi,err:=abi.JSON(strings.NewReader(abiStr))
	if err!=nil{
		return nil,err
	}
	commonAddress := common.HexToAddress(address)
	return &Contract{
		w, &abi, &commonAddress,
	},nil
}

func (c *Contract) Call(functionName string, args ...interface{}) (resp string,err error) {
	var dataByte []byte
	if args!=nil{
		dataByte,err=c.abi.Pack(functionName, args...)
	} else {
		dataByte = c.abi.Methods[functionName].ID()
	}
	if err!=nil{
		return
	}
	return c.web3.rpc.EthCall(T{
			To:c.address.String(),
			From:"0x0000000000000000000000000000000000000000",
			Data:fmt.Sprintf("0x%x", dataByte)},
			"latest",
			)
}

func (c *Contract) Send(params *SendTxParams, amount *big.Int, functionName string, args ...interface{}) (resp string,err error){
	if _,ok:=c.web3.privateKeyMap[strings.ToLower(params.fromAddress)];!ok{
		err=errors.New(fmt.Sprintf("from address %s not exist", params.fromAddress))
		return
	}

	data, err := c.abi.Pack(functionName, args...)
	if err != nil {
		return
	}

	tx := types.NewTransaction(
		params.nonce,
		*c.address,
		amount,
		params.gasLimit.Uint64(),
		params.gasPrice,
		data,
	)

	chainID := os.Getenv("CHAIN_ID")
	if len(chainID) == 0 {
		panic("need env CHAIN_ID")
	}
	rawData, err := utils.SignTx(c.web3.privateKeyMap[strings.ToLower(params.fromAddress)], os.Getenv("CHAIN_ID"), tx)
	if err != nil {
		panic(err)
	}
	return c.web3.rpc.EthSendRawTransaction(rawData)
}