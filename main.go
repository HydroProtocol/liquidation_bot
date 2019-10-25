package main

import (
	"auctionBidder/cli"
	"auctionBidder/client"
	"auctionBidder/utils"
	"auctionBidder/web3"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
)

func main() {
	var err error
	defer func() {
		if err != nil {
			spew.Dump(err)
		}
	}()

	setEnv()

	err = checkEnv()
	if err != nil {
		return
	}

	err = utils.InitDb()
	if err != nil {
		return
	}

	_, err = startBot()
	return
}

func startBot() (bot *cli.BidderBot, err error) {

	privateKey := os.Getenv("PRIVATE_KEY")
	maxSlippage, _ := decimal.NewFromString(os.Getenv("MAX_SLIPPAGE"))
	ethereumNodeUrl := os.Getenv("ETHEREUM_NODE_URL")
	minOrderValueUSD, _ := decimal.NewFromString(os.Getenv("MIN_ORDER_VALUE_USD"))
	gasPriceInGwei, _ := strconv.Atoi(os.Getenv("GAS_PRICE_TIPS_IN_GWEI"))
	profitMargin, _ := decimal.NewFromString(os.Getenv("PROFIT_MARGIN"))
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
		profitMargin,
	}

	go bot.Run()

	err = cli.StartGui()
	logrus.Error(err)

	return
}

func setEnv() {
	os.Setenv("CONFIGPATH", "/workingDir/config.json")
	os.Setenv("SQLITEPATH", "/workingDir/auctionBidderSqlite")
	os.Setenv("LOGPATH", "/workingDir")
	os.Setenv("CHAIN_ID", "1")
	os.Setenv("HYDRO_CONTRACT_ADDRESS", "0x241e82C79452F51fbfc89Fac6d912e021dB1a3B7")
	os.Setenv("DDEX_URL", "https://api.ddex.io/v4")

	// ropsten
	if os.Getenv("NETWORK") == "ropsten" {
		os.Setenv("CHAIN_ID", "3")
		os.Setenv("HYDRO_CONTRACT_ADDRESS", "0x06898143DF04616a8A8F9614deb3B99Ba12b3096")
		os.Setenv("DDEX_URL", "https://bfd-ropsten-59c1702d-api.intra.ddex.io/v4/")
	}

	loadEnv(os.Getenv("CONFIGPATH"))
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
	requiredEnvDefaultValue := map[string]string{
		"PRIVATE_KEY":            "B7A0C9D2786FC4DD080EA5D619D36771AEB0C8C26C290AFD3451B92BA2B7BC2C",
		"MAX_SLIPPAGE":           "0.05",
		"ETHEREUM_NODE_URL":      "https://mainnet.infura.io/v3/37851992caeb4289aa749112fe798621",
		"MIN_ORDER_VALUE_USD":    "100",
		"MARKETS":                "ETH-USDT,ETH-DAI",
		"PROFIT_MARGIN":          "0.01",
		"GAS_PRICE_TIPS_IN_GWEI": "5",
	}
	if os.Getenv("NETWORK") == "ropsten" {
		requiredEnvDefaultValue["ETHEREUM_NODE_URL"] = "https://ropsten.infura.io/v3/37851992caeb4289aa749112fe798621"
		requiredEnvDefaultValue["MARKETS"] = "ETH-USDT6,TETH-DAI"
	}
	for envName, defaultValue := range requiredEnvDefaultValue {
		if os.Getenv(envName) == "" {
			fmt.Printf("Enter %s(default %s):", envName, defaultValue)
			var input string
			fmt.Scanln(&input)
			if input == "" {
				input = defaultValue
			}
			os.Setenv(envName, input)
		}
		requiredEnvDefaultValue[envName] = os.Getenv(envName)
	}

	f, err := os.OpenFile(os.Getenv("CONFIGPATH"), os.O_CREATE|os.O_WRONLY, 0600)
	defer f.Close()
	if err == nil {
		envToWrite, _ := json.MarshalIndent(requiredEnvDefaultValue, "", "  ")
		f.Write(envToWrite)
	}

	return
}
