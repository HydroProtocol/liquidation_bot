package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"strings"
)

// hex string <-> int

func HexString2Int(str string) (int, error) {
	str = strings.ToLower(str)
	i, err := strconv.ParseInt(strings.TrimPrefix(str, "0x"), 16, 64)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

func Int2HexString(i int) string {
	return fmt.Sprintf("0x%x", i)
}

// hex string <-> big int

func HexString2BigInt(str string) (big.Int, error) {
	i := big.Int{}
	_, err := fmt.Sscan("0x"+strings.TrimPrefix(strings.ToLower(str), "0x"), &i)

	return i, err
}

func BigIntToHexString(bigInt big.Int) string {
	if bigInt.BitLen() == 0 {
		return "0x0"
	}

	return "0x" + strings.TrimPrefix(fmt.Sprintf("%x", bigInt.Bytes()), "0")
}

// hex string -> decimal

func HexString2Decimal(str string, exp int32) decimal.Decimal {
	i, _ := HexString2BigInt(str)
	return decimal.NewFromBigInt(&i, exp)
}

func String2Decimal(str string) decimal.Decimal {
	d, _ := decimal.NewFromString(str)
	return d
}

// bytes <-> hex string

func Bytes2HexString(bytes []byte) string {
	return "0x" + hex.EncodeToString(bytes)
}

func HexString2Bytes(str string) []byte {
	str = strings.TrimPrefix(strings.ToLower(str), "0x")

	if len(str)%2 == 1 {
		str = "0" + str
	}

	b, _ := hex.DecodeString(str)
	return b
}

// decimal <-> big int

func DecimalToBigInt(d decimal.Decimal) *big.Int {
	n := new(big.Int)
	n, _ = n.SetString(d.Floor().String(), 0)
	return n
}
