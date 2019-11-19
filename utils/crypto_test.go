package utils

import (
	"encoding/hex"
	"github.com/davecgh/go-spew/spew"
	"strings"
	"testing"
)

func TestPersonalSign(t *testing.T) {
	orderIdBytes, _ := hex.DecodeString(strings.TrimPrefix("0x3bfd186e2c45fb9fbfc8039906559d9e4181a5f1e8f45b05c6d673d3636bb949", "0x"))
	signature, _ := PersonalSign(orderIdBytes, "")
	spew.Dump("0x" + hex.EncodeToString(signature))
}
