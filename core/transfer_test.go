package core

import (
	"blocker/crypto"
	"blocker/types"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransferTransaction(t *testing.T) {
	// Make sure that transferTx bytes could send overnetwork
	fromPriv := crypto.GeneratePrivateKey()
	toPriv := crypto.GeneratePrivateKey()

	transfer := TransferTx{
		To: toPriv.Public().Address(),
	}
	assert.Nil(t, transfer.Sign(fromPriv))
	assert.NotNil(t, transfer.From)
	assert.NotNil(t, transfer.Signature)
	assert.Nil(t, transfer.Verify(nil))

	invalidOwner := crypto.GeneratePrivateKey()
	transfer.Signer = invalidOwner.Public()
	assert.NotNil(t, transfer.Verify(nil))
}

func TestTranaferTx(t *testing.T) {
	fromPriv := crypto.GeneratePrivateKey()
	toPriv := crypto.GeneratePrivateKey()

	transfer := TransferTx{
		To: toPriv.Public().Address(),
	}
	assert.Nil(t, transfer.Sign(fromPriv))
	assert.NotNil(t, transfer.From)
	assert.NotNil(t, transfer.Signature)
	assert.Nil(t, transfer.Verify(nil))

	buf := new(bytes.Buffer)

	assert.Nil(t, transfer.Encode(NewGobTransferTxEncoder(buf)))

	decodedTransfer := new(TransferTx)
	assert.Nil(t, decodedTransfer.Decode(NewGobTransferTxDecoder(buf)))

	assert.NotNil(t, decodedTransfer)

	invalidOwner := crypto.GeneratePrivateKey()
	transfer.Signer = invalidOwner.Public()
	assert.NotNil(t, transfer.Verify(nil))
}

func TestTxOutWithTxTransfer(t *testing.T) {
	fromPriv := crypto.GeneratePrivateKey()
	toPriv := crypto.GeneratePrivateKey()
	txOut := newTxOutForOwner(t, toPriv.Public(), 20)

	transfer := newTransferTx(t, fromPriv, toPriv.Public().Address(), 200, nil, []TxOut{txOut}, nil)

	buf := new(bytes.Buffer)
	assert.Nil(t, transfer.Encode(NewGobTransferTxEncoder(buf)))

	encodedTransfer := new(TransferTx)
	assert.Nil(t, encodedTransfer.Decode(NewGobTransferTxDecoder(buf)))

	assert.Equal(t, transfer, *encodedTransfer)

	invalidKey := crypto.GeneratePrivateKey()

	encodedTransfer.Signer = invalidKey.Public()

	assert.NotNil(t, encodedTransfer.Verify(nil))
}

func newTxOutForOwner(t *testing.T, to *crypto.PublicKey, value uint32) TxOut {
	scriptPub := crypto.ScriptPubKey(to)
	txOut := TxOut{
		ScriptPub: scriptPub,
		Value:     value,
	}
	return txOut
}

func newTransferTx(t *testing.T, from *crypto.PrivateKey, to types.Address, fee int32, txIn []TxIn, txOut []TxOut, prevTxx map[types.Hash]*Transaction) TransferTx {
	tx := TransferTx{
		To:  to,
		In:  txIn,
		Out: txOut,
		Fee: fee,
	}
	assert.Nil(t, tx.Sign(from))
	assert.NotNil(t, tx.Signature)
	assert.NotNil(t, tx.From)

	assert.Nil(t, tx.Verify(prevTxx))
	return tx
}
