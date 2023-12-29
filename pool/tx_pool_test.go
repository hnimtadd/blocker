package pool

import (
	"blocker/core"
	"blocker/types"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTxPoolAddSequenceTx(t *testing.T) {
	pool := NewTxPool(10)
	assert.Equal(t, 0, pool.PendingCount())
	tx := core.NewNativeTransaction([]byte("foo"))
	tx.Nonce = 0
	pool.Add(tx)
	tx = tx.Copy()
	tx.Nonce = 1
	pool.Add(tx)

	// pending should lock the pending pool until clearpending
	txx := pool.Pending()
	pool.LockPending()
	go func() {
		h := txx[0].Hash(core.TxHasher{})
		fmt.Println("deniding")
		pool.Denide([]types.Hash{h})
		fmt.Println("denided")
		pool.UnlockPending()
		txx = pool.Pending()
		pool.LockPending()
		assert.Equal(t, 1, len(txx))
		fmt.Println("...............")
		time.Sleep(time.Second * 1)
		pool.ClearPending()
	}()

	// this tx should not be remove with clearpending
	doneCh := make(chan struct{})
	go func() {
		tx = tx.Copy()
		tx.Nonce = 100
		fmt.Println("adding")
		pool.Add(tx) // this should block until clearpending
		fmt.Println("added")
		fmt.Println(pool.PendingCount())
		txx := pool.Pending()
		pool.LockPending()
		assert.Equal(t, 1, len(txx))
		pool.ClearPending()
		fmt.Println(pool.PendingCount())
		doneCh <- struct{}{}
	}()
	<-doneCh
}

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
