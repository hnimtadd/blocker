package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.Equal(t, len(privKey.Bytes()), privKeyLen)

	pubKey := privKey.Public()
	assert.Equal(t, len(pubKey.Bytes()), pubKeyLen)
}

func TestPrivateKeySign(t *testing.T) {
	privKey := GeneratePrivateKey()
	msg := []byte("foo bar baz")
	sig := privKey.Sign(msg)

	// Test ok
	assert.True(t, sig.Verify(privKey.Public(), msg))

	// Test with invalid msg
	assert.False(t, sig.Verify(privKey.Public(), []byte("fake message")))

	// Test with invalid publicKey
	invalidPrivKey := GeneratePrivateKey()
	assert.False(t, sig.Verify(invalidPrivKey.Public(), msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	address := pubKey.Address()
	assert.Equal(t, len(address.Bytes()), addressLen)
	fmt.Println(address)
}

func TestKeyFromString(t *testing.T) {
	var (
		seedString = "70e8b2282a89475436a50e13e94839b565f25d138eac87cbfee1bf3cca85d22d"
		privKey    = GeneratePrivateKeyFromString(seedString)
		pubKey     = privKey.Public()
		address    = "0393f29f09c56a1d108a3ba1a9adbba889eddaa1"
	)
	// fmt.Println(pubKey.Address())
	assert.Equal(t, address, pubKey.Address().String())
}

func TestKeyFromSeed(t *testing.T) {
	var (
		seedString = "70e8b2282a89475436a50e13e94839b565f25d138eac87cbfee1bf3cca85d22d"
	)
	seed, err := hex.DecodeString(seedString)
	assert.NoError(t, err)
	assert.NotEqual(t, len(seed), 0)

	var (
		privKey = GeneratePrivateKeyWithSeed(seed)
		pubKey  = privKey.Public()
		address = "0393f29f09c56a1d108a3ba1a9adbba889eddaa1"
	)
	assert.Equal(t, address, pubKey.Address().String())

}
