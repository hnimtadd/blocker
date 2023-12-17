package core

import (
	"blocker/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryCollection(t *testing.T) {
	repo := NewInMemoryStorage()
	tx := NewNativeTransaction(nil)
	hash := tx.Hash(TxHasher{})
	assert.Nil(t, repo.PutCollection(tx))
	gtx, err := repo.GetCollection(hash)
	assert.Nil(t, err)
	assert.Equal(t, *gtx, *tx)
	assert.Equal(t, ErrExisted, repo.PutCollection(tx))

	ntx, err := repo.GetCollection(types.RandomHash())
	assert.Equal(t, ErrNotExisted, err)
	assert.Nil(t, ntx)

	// txx, err := repo.GetAll()
	// assert.Nil(t, err)
	//
	// assert.Equal(t, 1, len(txx))
	// assert.Equal(t, gtx, txx[hash])
}
