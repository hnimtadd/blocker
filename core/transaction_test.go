package core

import (
	"blocker/crypto"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignTransaction(t *testing.T) {
	data := []byte("sample")
	tx := Transaction{
		Data: data,
	}

	var (
		privKey = crypto.GeneratePrivateKey()
		pubKey  = privKey.Public()
	)
	tx.From = pubKey

	assert.NotNil(t, tx.Verify())

	assert.Nil(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Signature)
	assert.NotNil(t, tx.From)
	assert.Equal(t, tx.From, pubKey)
	assert.Nil(t, tx.Verify())
}

func TestSignInvalidTransaction(t *testing.T) {
	data := []byte("sample")
	tx := Transaction{
		Data: data,
	}

	var (
		privKey        = crypto.GeneratePrivateKey()
		pubKey         = privKey.Public()
		invalidPrivKey = crypto.GeneratePrivateKey()
		invalidPubKey  = invalidPrivKey.Public()
	)

	assert.Nil(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Signature)
	assert.NotNil(t, tx.From)
	assert.Equal(t, tx.From, pubKey)
	assert.Nil(t, tx.Verify())

	tx.From = invalidPubKey
	assert.NotNil(t, tx.Verify())
}

func TestEncodeDecodeTransaction(t *testing.T) {
	data := []byte("sample")
	tx := Transaction{
		Data: data,
	}
	buf := &bytes.Buffer{}

	enc := &GobTxEncoder{
		w: buf,
	}
	assert.Nil(t, tx.Encode(enc))

	txDecode := Transaction{}
	dnc := &GobTxDecoder{
		r: buf,
	}
	assert.Nil(t, txDecode.Decode(dnc))

	assert.Equal(t, tx, txDecode)
}

func RandomTx(t *testing.T) *Transaction {
	tx := &Transaction{
		Data: []byte("Foo"),
	}
	return tx
}

func RandomTxWithSignature(t *testing.T) *Transaction {
	tx := &Transaction{
		Data: []byte("Foo"),
	}
	privKkey := crypto.GeneratePrivateKey()
	assert.Nil(t, tx.Sign(privKkey))
	return tx
}
