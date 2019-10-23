package main

import (
	"auctionBidder/cli"
	"auctionBidder/client"
	"auctionBidder/utils"
	"auctionBidder/web3"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
)

func main() {
	var err error
	defer spew.Dump(err)
	err = loadEnv("./env.json")
	if err != nil {
		return
	}

	err = loadEnv(os.Getenv("CONFIGPATH"))
	if err != nil {
		return
	}

	err = checkEnv()
	if err != nil {
		return
	}

	_, err = startBot()
	return
}

func startBot() (bot *cli.BidderBot, err error) {

	err = utils.InitDb()
	if err != nil {
		return
	}

	privateKey := os.Getenv("PRIVATE_KEY")
	maxSlippage, _ := decimal.NewFromString(os.Getenv("MAX_SLIPPAGE"))
	ethereumNodeUrl := os.Getenv("ETHEREUM_NODE_URL")
	minOrderValueUSD, _ := decimal.NewFromString(os.Getenv("MIN_ORDER_VALUE_USD"))
	gasPriceInGwei, _ := strconv.Atoi(os.Getenv("GAS_PRICE_TIPS_IN_GWEI"))
	profitBuffer, _ := decimal.NewFromString(os.Getenv("PROFIT_BUFFER"))
	markets := os.Getenv("MARKETS")

	ddexClient, err := client.NewDdexClient(privateKey)

	bidderClient, err := client.NewBidderClient(privateKey, ddexClient.Assets, ddexClient.Markets)
	if err != nil {
		return
	}

	web3Client := web3.NewWeb3(ethereumNodeUrl)

	bot = &cli.BidderBot{
		bidderClient,
		ddexClient,
		web3Client.NewBlockChannel(),
		maxSlippage,
		minOrderValueUSD,
		gasPriceInGwei,
		markets,
		profitBuffer,
	}

	go bot.Run()

	err = cli.StartGui()
	logrus.Error(err)

	return
}

func loadEnv(filePath string) (err error) {
	jsonFile, err := os.Open(filePath)
	defer jsonFile.Close()
	if err != nil {
		return
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var envs map[string]string
	json.Unmarshal(byteValue, &envs)
	for key, value := range envs {
		os.Setenv(key, value)
	}
	return
}

func checkEnv() (err error) {
	requiredEnv := []string{
		"PRIVATE_KEY",
		"MAX_SLIPPAGE",
		"ETHEREUM_NODE_URL",
		"CHAIN_ID",
		"HYDRO_CONTRACT_ADDRESS",
		"MIN_AMOUNT_VALUE_USD",
		"DDEX_URL",
		"MARKETS",
		"PROFIT_BUFFER",
	}
	for _, envName := range requiredEnv {
		if os.Getenv(envName) == "" {
			logrus.Panicf("environment variable %s missed", envName)
			// fmt.Printf("Enter %s:", envName)
			// var input string
			// fmt.Scanln(&input)
			// os.Setenv(envName, input)
		}
	}
	return
}

func prepareAccount(web3Client *web3.Web3, tokenAddress string) (err error) {
	// TODO AUTO APPROVE AND DEPOSIT
	// BE AWARE OF USER SETTING WRONG HYDRO CONTRACT ADDRESS
	return
}
