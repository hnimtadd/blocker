package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasher(t *testing.T) {
	tx := NewNativeTransaction([]byte("hello world"))
	txHash := tx.Hash(TxHasher{})
	println(tx.String())

	tx.Nonce = 1
	newTxHash := tx.ReHash(TxHasher{})
	println(tx.String())
	assert.NotEqual(t, txHash, newTxHash)
}
