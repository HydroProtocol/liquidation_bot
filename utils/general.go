package utils

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
	"math/rand"
	"time"
)

func MillisecondTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func SetDecimal(d decimal.Decimal, decimalNum int) decimal.Decimal {
	return d.Truncate(int32(decimalNum))
}

func SetPrecision(d decimal.Decimal, precision int) decimal.Decimal {
	if precision <= 0 {
		panic("precision must greater than 0")
	}
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
	validDecimal, err := decimal.NewFromString(numString[:endPosition])
	if err != nil {
		panic("set precision failed")
	}
	return validDecimal
}

func ToggleSide(side string) string {
	if side == SELL {
		return BUY
	} else {
		return SELL
	}
}

func GetUniqueId() string {
	timestamp := time.Now().String()
	randomNum := rand.Intn(10000)
	s := spew.Sprintf("%s@%d", timestamp, randomNum)
	h := md5.New()
	h.Write([]byte(s))
	sha1Hash := hex.EncodeToString(h.Sum(nil))
	return sha1Hash
}

