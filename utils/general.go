package utils

import (
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

func MillisecondTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func SetDecimal(d decimal.Decimal, decimalNum int) decimal.Decimal {
	return d.Truncate(int32(decimalNum))
}

func SetPrecision(d decimal.Decimal, precision int) decimal.Decimal {
	numString := d.String()
	precisionCount := 0
	endPosition := 0
	for _, c := range numString {
		if c != '.' {
			precisionCount += 1
		}
		if precisionCount > precision {
			break
		}
		endPosition += 1
	}
	validDecimal, _ := decimal.NewFromString(numString[:endPosition])
	return validDecimal
}

func IsAddressEqual(a string, b string) bool {
	a = strings.TrimPrefix(strings.ToLower(a), "0x")
	b = strings.TrimPrefix(strings.ToLower(b), "0x")
	return a == b
}
