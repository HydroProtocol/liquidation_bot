package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"strings"
)

// ParseInt parse hex string value to int
func ParseInt(value string) (int, error) {
	i, err := strconv.ParseInt(strings.TrimPrefix(value, "0x"), 16, 64)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

// ParseBigInt parse hex string value to big.Int
func ParseBigInt(value string) (big.Int, error) {
	i := big.Int{}
	_, err := fmt.Sscan(value, &i)

	return i, err
}

// IntToHex convert int to hexadecimal representation
func IntToHex(i int) string {
	return fmt.Sprintf("0x%x", i)
}

// BigToHex covert big.Int to hexadecimal representation
func BigToHex(bigInt big.Int) string {
	if bigInt.BitLen() == 0 {
		return "0x0"
	}

	return "0x" + strings.TrimPrefix(fmt.Sprintf("%x", bigInt.Bytes()), "0")
}

func Int2Hex(number uint64) string {
	return fmt.Sprintf("%x", number)
}

// just return uint64 type
func Hex2Int(hex string) uint64 {
	if strings.HasPrefix(hex, "0x") || strings.HasPrefix(hex, "0X") {
		hex = hex[2:]
	}
	intNumber, err := strconv.ParseUint(hex, 16, 64)

	if err != nil {
		return 0
	}

	return uint64(intNumber)
}

func Bytes2Hex(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func Hex2Bytes(str string) []byte {
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		str = str[2:]
	}

	if len(str)%2 == 1 {
		str = "0" + str
	}

	h, _ := hex.DecodeString(str)
	return h
}

// with prefix '0x'
func Bytes2HexP(bytes []byte) string {
	return "0x" + hex.EncodeToString(bytes)
}

func Hex2BigInt(str string) *big.Int {
	bytes := Hex2Bytes(str)
	b := big.NewInt(0)
	b.SetBytes(bytes)
	return b
}

func Bytes2BigInt(bytes []byte) *big.Int {
	b := big.NewInt(0)
	b.SetBytes(bytes)
	return b
}

// RightPadBytes zero-pads slice to the right up to length l.
func RightPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded, slice)

	return padded
}

// LeftPadBytes zero-pads slice to the left up to length l.
func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

func Int2Bytes(i uint64) []byte {
	return Hex2Bytes(Int2Hex(i))
}

func DecimalToBigInt(d decimal.Decimal) *big.Int {
	n := new(big.Int)
	n, _ = n.SetString(d.String(), 0)
	return n
}