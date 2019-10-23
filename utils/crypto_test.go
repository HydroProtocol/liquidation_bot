package utils

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPrivateKey(t *testing.T) {
	address := "0x93388b4efe13b9b18ed480783c05462409851547"
	prvKeyHex := "95b0a982c0dfc5ab70bf915dcf9f4b790544d25bc5e6cff0f38a59d0bba58651"
	except := address

	act, _ := NewPrivateKeyByHex(prvKeyHex)
	assert.EqualValues(t, except, PubKey2Address(act.PublicKey))

	prvKeyHex2 := "95b0a982c0dfc5ab70bf915dcf9f4b790544d25bc5e6cff0f38a59d0bba586"
	act2, err := NewPrivateKeyByHex(prvKeyHex2)
	assert.Nil(t, act2)
	assert.EqualValues(t, err.Error(), "invalid length, need 256 bits")

	prvKeyHex3 := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	act3, err := NewPrivateKeyByHex(prvKeyHex3)
	assert.Nil(t, act3)
	assert.EqualValues(t, err.Error(), "invalid private key, >=N")

	prvKeyHex4 := "0000000000000000000000000000000000000000000000000000000000000000"
	act4, err := NewPrivateKeyByHex(prvKeyHex4)
	assert.Nil(t, act4)
	assert.EqualValues(t, err.Error(), "invalid private key, zero or negative")

}

func TestNewPrivateKeyByHex(t *testing.T) {
	address := "0x93388b4efe13b9b18ed480783c05462409851547"
	prvKeyHex := "95b0a982c0dfc5ab70bf915dcf9f4b790544d25bc5e6cff0f38a59d0bba58651"
	prvKeyBytes := HexString2Bytes(prvKeyHex)
	except := address

	act, _ := NewPrivateKey(prvKeyBytes)

	assert.EqualValues(t, except, PubKey2Address(act.PublicKey))
}

func TestSign(t *testing.T) {
	prvKeyHex := "95b0a982c0dfc5ab70bf915dcf9f4b790544d25bc5e6cff0f38a59d0bba58651"
	message := HexString2Bytes("9df8dba3720d00bd48ad744722021ef91b035e273bccfb78660ca8df9574b086")
	except := HexString2Bytes("2736b2ca3e2d4778e53a33e0d9bb2d9bad91ec858ab71ad49e31f540f15728a83dbea28bd686bb66d06e4ad9f48912ef437b92a272ea47563c2df80ed59b508e00")
	actKey, _ := NewPrivateKeyByHex(prvKeyHex)
	act, _ := Sign([]byte(message), actKey)
	assert.EqualValues(t, except, act)
}

func TestPersonalSignAndPersonalSignByPrivateKey(t *testing.T) {
	prvKeyHex := "95b0a982c0dfc5ab70bf915dcf9f4b790544d25bc5e6cff0f38a59d0bba58651"
	message := HexString2Bytes("9df8dba3720d00bd48ad744722021ef91b035e273bccfb78660ca8df9574b086")
	except := HexString2Bytes("aa7cd9f5a7eb485771215d45cc2a4c535e270c75c3595ae6b1c158aef72e67066ad5df037ad5945c65da90edfaa4fe418e5b6bd2225ec9d4b704433a779e4bff00")

	actKey, _ := NewPrivateKeyByHex(prvKeyHex)
	act, _ := PersonalSignByPrivateKey(message, actKey)
	act2, _ := PersonalSign(message, prvKeyHex)
	assert.EqualValues(t, except, act)
	assert.EqualValues(t, except, act2)
}

func TestPersonalSignAndPersonalSignByPrivateKey1(t *testing.T) {
	prvKeyHex := "95b0a982c0dfc5ab70bf915dcf9f4b790544d25bc5e6cff0f38a59d0bba58651"
	message := HexString2Bytes("0x35fa76fab0816f41f7a5fd6b800bd13394e6ce1c39af9e79cc372a92a3c44c2a")

	actKey, _ := NewPrivateKeyByHex(prvKeyHex)
	act, _ := PersonalSignByPrivateKey(message, actKey)
	act2, _ := PersonalSign(message, prvKeyHex)

	spew.Dump(Bytes2HexString(act))
	spew.Dump(Bytes2HexString(act2))
}

func TestEcRecover(t *testing.T) {
	sign := HexString2Bytes("2736b2ca3e2d4778e53a33e0d9bb2d9bad91ec858ab71ad49e31f540f15728a83dbea28bd686bb66d06e4ad9f48912ef437b92a272ea47563c2df80ed59b508e00")
	message := Keccak256([]byte("some message"))
	except := HexString2Bytes("0450d7aa97f7496fd412f393e54df0cbe3f6cbeacf15d1ddb12133e408522feb8896dd1652ee84b18788bc7753663302a6489f779352bbfec010ab25c9e3806843")

	act, _ := EcRecover(message, sign)
	assert.EqualValues(t, except, act)
}

func TestPersonalEcRecover(t *testing.T) {
	address := "0x93388b4efe13b9b18ed480783c05462409851547"
	sign := HexString2Bytes("aa7cd9f5a7eb485771215d45cc2a4c535e270c75c3595ae6b1c158aef72e67066ad5df037ad5945c65da90edfaa4fe418e5b6bd2225ec9d4b704433a779e4bff00")
	message := Keccak256([]byte("some message"))

	act, _ := PersonalEcRecover(message, sign)
	assert.EqualValues(t, address[2:], act)
}

func TestSigToPub(t *testing.T) {
	address := "0x93388b4efe13b9b18ed480783c05462409851547"
	message := Keccak256([]byte("some message"))
	sign := HexString2Bytes("aa7cd9f5a7eb485771215d45cc2a4c535e270c75c3595ae6b1c158aef72e67066ad5df037ad5945c65da90edfaa4fe418e5b6bd2225ec9d4b704433a779e4bff00")

	act, _ := SigToPub(hashPersonalMessage(message), sign)
	assert.EqualValues(t, address, PubKey2Address(*act))
}
