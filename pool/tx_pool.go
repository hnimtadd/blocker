package pool

import (
	"blocker/core"
	"blocker/types"
	"errors"
	"fmt"
	"sync"
	"time"
)

type TxPool struct {
	pending   *TxSortedMap
	processed *TxSortedMap
	denided   *TxSortedMap
	expired   *TxSortedMap
	maxLength int // The maxLength of the total pool of transactions. When the pool is full the oldest transaction will be pruned.
}
type TxPoolStatus string

const (
	TxPooldDenied  TxPoolStatus = "denied"
	TxPoolExpired  TxPoolStatus = "expired"
	TxPooldPending TxPoolStatus = "pending"
	TxPoolReceived TxPoolStatus = "received"
	TxPoolUnknown  TxPoolStatus = "unknown"
)

var (
	ErrMaxLengthExceed = errors.New("max length exceed")
	ErrNotknown        = errors.New("unknown error occurs")
)

func NewTxPool(maxLength int) *TxPool {
	return &TxPool{
		processed: NewTxSortedMap(),
		pending:   NewTxSortedMap(),
		denided:   NewTxSortedMap(),
		expired:   NewTxSortedMap(),
		maxLength: maxLength,
	}
}

func (p *TxPool) Add(tx *core.Transaction) error {
	// prune the oldest transaction that is sitting in the all pool
	if p.pending.Count() == p.maxLength {
		return ErrMaxLengthExceed
	}
	if !p.pending.Contains(tx.ReHash(core.TxHasher{})) {
		p.pending.Add(tx)
		return nil
	}
	return ErrNotknown
}

// Contains check if transactions are currently in pool pending pool
func (p *TxPool) Contains(hash types.Hash) bool {
	return p.pending.Contains(hash)
}

/*
Get get transactions from pool and transaction status

- If transaction are in comfirmed pool, that mean the tx is already in the blockchain

- If transaction are in pending pool, that mean the tx is received by the node

- If transaction are in denided pool, that mean the tx is already denided by the blockchain
*/
func (p *TxPool) Get(hash types.Hash) (TxPoolStatus, *core.Transaction, error) {
	tx := p.pending.Get(hash)
	if tx != nil {
		return TxPoolReceived, tx, nil
	}
	tx = p.denided.Get(hash)
	if tx != nil {
		return TxPooldDenied, tx, nil
	}
	tx = p.processed.Get(hash)
	if tx != nil {
		return TxPooldPending, tx, nil
	}
	return TxPoolUnknown, tx, nil
}

func (p *TxPool) LockPending() {
	p.pending.Lock()
}

func (p *TxPool) UnlockPending() {
	p.pending.Unlock()
}

// Pending returns a slice of transactions that are in the pending pool, and then lock the pending
func (p *TxPool) Pending() []*core.Transaction {
	txx := p.pending.txx.Data
	p.LockPending()
	now := time.Now().UnixNano()
	expiredTx := []types.Hash{}
	pendingTxx := []*core.Transaction{}
	for _, tx := range txx {
		if tx.ValidUntil < now && tx.ValidUntil != 0 {
			expiredTx = append(expiredTx, tx.Hash(core.TxHasher{}))
			continue
		}
		if tx.ValidFrom == 0 {
			pendingTxx = append(pendingTxx, tx)
			continue
		}
		fmt.Println(tx)
		if tx.ValidFrom < now {
			pendingTxx = append(pendingTxx, tx)
			continue
		}
	}
	if len(expiredTx) > 0 {
		p.Expire(expiredTx)
		return pendingTxx
	}
	p.UnlockPending()
	return pendingTxx
}

func (p *TxPool) Expire(idxx []types.Hash) {
	p.UnlockPending()
	for _, hash := range idxx {
		tx := p.pending.Remove(hash)
		p.expired.Add(tx)
		// fmt.Println(tx, "expired")
	}
	p.LockPending()
}

func (p *TxPool) ClearPending() {
	p.pending.Unlock()
	p.pending.Clear()
	// fmt.Println("clear")
}

func (p *TxPool) PendingCount() int {
	return p.pending.Count()
}

// Denide unlock the locked pending, remove transactions from the current pending, and lock pending again, this method must run after pending
func (p *TxPool) Denide(idxx []types.Hash) []*core.Transaction {
	p.pending.Unlock()
	denidedTXX := []*core.Transaction{}
	for _, idx := range idxx {
		tx := p.pending.Remove(idx)
		denidedTXX = append(denidedTXX, tx)
		p.denided.Add(tx)
	}
	p.pending.Lock()
	return denidedTXX
}

func (p *TxPool) Processed(txx []*core.Transaction) {
	p.pending.Unlock()
	for _, tx := range txx {
		p.processed.Add(tx)
		p.pending.Remove(tx.Hash(core.TxHasher{}))
	}
}

type TxSortedMap struct {
	lookup map[types.Hash]*core.Transaction
	txx    *types.List[*core.Transaction]
	lock   sync.RWMutex
}

func NewTxSortedMap() *TxSortedMap {
	return &TxSortedMap{
		lookup: make(map[types.Hash]*core.Transaction),
		txx:    types.NewList[*core.Transaction](),
	}
}

func (t *TxSortedMap) First() *core.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	first := t.txx.Get(0)
	return t.lookup[first.Hash(core.TxHasher{})]
}

func (t *TxSortedMap) Get(h types.Hash) *core.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.lookup[h]
}

func (t *TxSortedMap) Add(tx *core.Transaction) {
	hash := tx.ReHash(core.TxHasher{})

	t.lock.Lock()
	defer t.lock.Unlock()

	if _, ok := t.lookup[hash]; !ok {
		t.lookup[hash] = tx
		t.txx.Insert(tx)
	}
}

func (t *TxSortedMap) Remove(h types.Hash) *core.Transaction {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.txx.Remove(t.lookup[h])
	tx := t.lookup[h]
	delete(t.lookup, h)
	return tx
}

func (t *TxSortedMap) Count() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.lookup)
}

func (t *TxSortedMap) Contains(h types.Hash) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	_, ok := t.lookup[h]
	return ok
}

func (t *TxSortedMap) Clear() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.lookup = make(map[types.Hash]*core.Transaction)
	t.txx.Clear()
}

func (t *TxSortedMap) Lock() {
	t.lock.Lock()
}

func (t *TxSortedMap) Unlock() {
	t.lock.Unlock()
}
