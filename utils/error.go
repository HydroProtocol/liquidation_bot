package utils

import "fmt"

var (
	AddressNotExist         = fmt.Errorf("address not exist")
	AssetNotExist           = fmt.Errorf("asset not exist")
	OrderbookNotComplete    = fmt.Errorf("orderbook not complete")
	OrderbookDepthNotEnough = fmt.Errorf("orderbook depth not enough")
	TransactionFailed       = fmt.Errorf("transaction failed")
	TransactionNotFound     = fmt.Errorf("transaction not found")
	AuctionNotExist         = fmt.Errorf("auction not exist")
)
