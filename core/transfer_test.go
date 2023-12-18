package core

import (
	"blocker/crypto"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testTransferTx(t *testing.T) {
	from := crypto.GeneratePrivateKey()
	to := crypto.GeneratePrivateKey()
	transfer := TransferTx{
		To:    to.Public(),
		Value: 100,
	}

	assert.Nil(t, transfer.Sign(from))
	assert.NotNil(t, transfer.From)
	assert.NotNil(t, transfer.Signature)

	invalidFrom := crypto.GeneratePrivateKey()

	transfer.From = invalidFrom.Public()
	assert.NotNil(t, transfer.Verify())

	assert.Nil(t, transfer.Verify())

	buf := new(bytes.Buffer)

	assert.Nil(t, transfer.Encode(NewGobTransferTxEncoder(buf)))

	decodedTransferTx := new(TransferTx)

	assert.Nil(t, decodedTransferTx.Decode(NewGobTransferTxDecoder(buf)))
}
