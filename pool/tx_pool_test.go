package pool

import (
	"blocker/core"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxPool(t *testing.T) {
	pool := NewTxPool(10)
	assert.Equal(t, 0, pool.PendingCount())
}

func TestTxPoolAddTx(t *testing.T) {
	p := NewTxPool(10)
	assert.Equal(t, 0, p.PendingCount())
	tx := core.NewNativeTransaction([]byte("foo"))
	p.Add(tx)
	assert.Equal(t, 1, p.PendingCount())
}

func TestTxPoolFlush(t *testing.T) {
	p := NewTxPool(10)
	assert.Equal(t, 0, p.PendingCount())
	tx := core.NewNativeTransaction([]byte("foo"))
	p.Add(tx)
	assert.Equal(t, 1, p.PendingCount())
	p.ClearPending()
	assert.Equal(t, 0, p.PendingCount())
	tx2 := core.NewNativeTransaction([]byte("new"))
	p.Add(tx2)
	assert.Equal(t, 1, p.PendingCount())
}

func TestTxPoolAddDuplicateTx(t *testing.T) {
	p := NewTxPool(10)
	assert.Equal(t, 0, p.PendingCount())
	tx := core.NewNativeTransaction([]byte("foo"))
	p.Add(tx)
	assert.Equal(t, 1, p.PendingCount())
	p.Add(tx)
	assert.Equal(t, 1, p.PendingCount())
}

func TestTxPoolTransactions(t *testing.T) {
	p := NewTxPool(10)
	assert.Equal(t, 0, p.PendingCount())
	txLen := 1000
	for i := 0; i < txLen; i++ {
		tx := core.NewNativeTransaction([]byte(fmt.Sprintf("%v", i)))
		tx.SetTimestamp(int64(i))
		p.Add(tx)
	}
	txx := p.Pending()
	for i := 0; i < len(txx)-1; i++ {
		assert.NotNil(t, txx[i])
		assert.NotNil(t, txx[i+1])
		assert.LessOrEqual(t, txx[i].Timestamp(), txx[i+1].Timestamp())
	}
}
